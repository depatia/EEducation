package auth

import (
	"AuthService/internal/models"
	serviceerrors "AuthService/internal/services/service_errors"
	"AuthService/internal/storage/storage"
	"AuthService/pkg/tools/hash"
	"AuthService/pkg/tools/jwt"
	"AuthService/pkg/tools/logger/sl"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
)

type AuthStore struct {
	jwt              jwt.JwtWrapper
	userCreater      UserCreater
	userProvider     UserProvider
	permissionSetter PermissionSetter
	permissionGetter PermissionGetter
	log              *slog.Logger
}

func New(
	jwt jwt.JwtWrapper,
	userCreater UserCreater,
	userProvider UserProvider,
	permissionSetter PermissionSetter,
	permissionGetter PermissionGetter,
	log *slog.Logger,
) *AuthStore {
	return &AuthStore{
		jwt:              jwt,
		userCreater:      userCreater,
		userProvider:     userProvider,
		permissionSetter: permissionSetter,
		permissionGetter: permissionGetter,
		log:              log,
	}
}

type UserCreater interface {
	CreateUser(ctx context.Context, email string, hash []byte) (int64, error)
}

type UserProvider interface {
	GetUser(ctx context.Context, email string) (models.User, error)
	UpdatePassword(ctx context.Context, userID int64, passHash []byte) error
}

type PermissionSetter interface {
	SetPermission(ctx context.Context, userID int64, permissionLevel int64) error
}

type PermissionGetter interface {
	GetPermission(ctx context.Context, userID int64) (int64, error)
}

func (a *AuthStore) RegisterUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.Register"

	log := a.log.With(
		slog.String("Operation", op),
		slog.String("Username", email),
	)

	log.Info("register user")

	_, err := mail.ParseAddress(email)
	if err != nil {
		log.Error("failed to register", sl.Err(err))
		return 0, fmt.Errorf("failed to create user due to error: %w", serviceerrors.ErrBadEmailFormat)
	}

	passHash, err := hash.HashPass(pass)
	if err != nil {
		log.Error("failed to register", sl.Err(err))

		return 0, fmt.Errorf("failed to generate hash password due to error: %w", err)
	}

	id, err := a.userCreater.CreateUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to register", sl.Err(err))

		return 0, fmt.Errorf("failed to create user due to error: %w", err)
	}

	log.Info("user is registered")

	return id, nil
}

func (a *AuthStore) Login(ctx context.Context, email string, password string) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("Operation", op),
		slog.String("Username", email),
	)

	log.Info("login user")

	user, err := a.userProvider.GetUser(ctx, email)

	if err != nil {
		log.Error("failed to login", sl.Err(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return "", fmt.Errorf("user not found due to error: %w", storage.ErrUserNotFound)
		}
		return "", fmt.Errorf("failed to get user due to error: %w", err)
	}

	if ok := hash.CheckPass(password, []byte(user.PassHash)); !ok {
		log.Error("failed to login", sl.Err(serviceerrors.ErrInvalidCredentials))

		return "", fmt.Errorf("invalid credentials due to error: %w", serviceerrors.ErrInvalidCredentials)
	}

	token, err := a.jwt.NewToken(user)
	if err != nil {
		log.Error("failed to login", sl.Err(err))

		return "", fmt.Errorf("failed to generate token due to error: %w", err)
	}

	log.Info("user logged in")

	return token, nil
}

func (a *AuthStore) ChangePassword(ctx context.Context, email, oldPassword, newPassword, token string) error {
	const op = "auth.ChangePassword"

	log := a.log.With(
		slog.String("Operation", op),
		slog.String("Username", email),
	)

	log.Info("updating user password")

	user, err := a.userProvider.GetUser(ctx, email)
	if err != nil {
		log.Error("failed to change password", sl.Err(err))

		return err
	}

	if ok := hash.CheckPass(oldPassword, []byte(user.PassHash)); !ok {
		log.Error("failed to change password", sl.Err(serviceerrors.ErrInvalidCredentials))

		return serviceerrors.ErrInvalidCredentials
	}

	_, err = a.Validate(ctx, token)
	if err != nil {
		log.Error("failed to change password", sl.Err(err))

		return err
	}

	passHash, err := hash.HashPass(newPassword)
	if err != nil {
		log.Error("failed to change password", sl.Err(err))

		return err
	}

	if err = a.userProvider.UpdatePassword(ctx, user.ID, passHash); err != nil {
		log.Error("failed to change password", sl.Err(err))

		return err
	}

	log.Info("user password successfully changed")

	return nil
}

func (a *AuthStore) Validate(ctx context.Context, token string) (int64, error) {
	const op = "auth.Validate"

	log := a.log.With(
		slog.String("Operation", op),
		slog.String("Token", token),
	)

	log.Info("validating jwt token")

	var u models.User

	claims, err := a.jwt.ValidateToken(token)

	if err != nil {
		log.Error("failed to validate jwt", sl.Err(err))

		return 0, err
	}

	u, err = a.userProvider.GetUser(ctx, claims.Email)

	if err != nil {
		log.Error("failed to validate jwt", sl.Err(err))

		return 0, err
	}

	log.Info("jwt token is valid")

	return u.ID, nil
}

func (a *AuthStore) SetPermissionLevel(ctx context.Context, userID, permissionLevel, initiatorID int64) error {
	const op = "auth.SetPermissionLevel"

	log := a.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("updating user permissions")

	lvl, err := a.permissionGetter.GetPermission(ctx, initiatorID)
	if err != nil {
		log.Error("failed to change permissions", sl.Err(err))

		return err
	}

	if lvl < 3 {
		log.Error("failed to change permissions", sl.Err(serviceerrors.ErrAccessDenied))

		return serviceerrors.ErrAccessDenied
	}

	if err = a.permissionSetter.SetPermission(ctx, userID, permissionLevel); err != nil {
		log.Error("failed to change permissions", sl.Err(err))

		return err
	}

	log.Info("user permissions updated")

	return nil
}

func (a *AuthStore) GetPermissionLevel(ctx context.Context, userID int64) (int64, error) {
	const op = "auth.GetPermissionLevel"

	log := a.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("getting user permissions")

	lvl, err := a.permissionGetter.GetPermission(ctx, userID)
	if err != nil {
		log.Error("failed to get permissions", sl.Err(err))

		return 0, err
	}

	log.Info("user permissions given")

	return lvl, nil
}
