package user

import (
	clienterrors "APIGateway/internal/client_errors"
	"APIGateway/pkg/tools/logger/sl"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/depatia/EEducation-Protos/gen/user"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api user.UserServiceClient
	log *slog.Logger
}

type AuthService interface {
	Register(ctx context.Context, req *RegisterReq) (int, error)
	Login(ctx context.Context, req *LoginReq) (string, int, error)
	UpdatePassword(ctx context.Context, req *UpdatePasswordReq) (int, error)
	SetPermissionLevel(ctx context.Context, req *SetPermissionLevelReq) (int, error)
	GetPermissionLevel(ctx context.Context, userID string) (int64, error)
}

type UserService interface {
	FillUserProfile(ctx context.Context, req *FillUserProfileReq) (int, error)
	ChangeUserStatus(ctx context.Context, req *ChangeUserStatusReq) (int, error)
	IsUserActive(ctx context.Context, req *IsUserActiveReq) (bool, int, error)
	GetStudentsByClassname(ctx context.Context, classname string) ([]*user.Student, int, error)
	DeleteUser(ctx context.Context, userID string) (int, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := user.NewUserServiceClient(cc)

	return &Client{
		api: grpcClient,
		log: log,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Register(ctx context.Context, req *RegisterReq) (int, error) {
	const op = "grpc.auth.Register"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Username", req.Email),
	)

	log.Info("register user")

	if req.Email == "" || req.Password == "" {
		log.Error("failed to register", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	_, err := c.api.Register(ctx, &user.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		log.Error("failed to register", sl.Err(err))

		if status.Code(err) == codes.AlreadyExists {
			return http.StatusConflict, fmt.Errorf("%s: %w", op, clienterrors.ErrAlreadyExists)
		}
		if status.Code(err) == codes.InvalidArgument {
			return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrBadEmailFormat)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user is registered")

	return http.StatusCreated, nil
}

func (c *Client) Login(ctx context.Context, req *LoginReq) (string, int, error) {
	const op = "grpc.auth.Login"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Username", req.Email),
	)

	log.Info("login user")

	if req.Email == "" || req.Password == "" {
		log.Error("failed to login", sl.Err(clienterrors.ErrAllFieldsRequired))

		return "", http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.Login(ctx, &user.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Error("failed to login", sl.Err(err))

		if status.Code(err) == codes.Unauthenticated {
			return "", http.StatusUnauthorized, fmt.Errorf("%s: %w", op, clienterrors.ErrIncorrectCredentials)
		}
		if status.Code(err) == codes.NotFound {
			return "", http.StatusNotFound, fmt.Errorf("%s: %w", op, clienterrors.ErrUserNotFound)
		}
		return "", http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	return resp.Token, http.StatusOK, nil
}

func (c *Client) UpdatePassword(ctx context.Context, req *UpdatePasswordReq) (int, error) {
	const op = "grpc.auth.UpdatePassword"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Username", req.Email),
	)

	log.Info("updating user password")

	if req.Email == "" || req.OldPassword == "" || req.NewPassword == "" || req.Token == "" {
		log.Error("failed to change password", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.UpdatePassword(ctx, &user.UpdatePasswordRequest{
		Email:       req.Email,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		Token:       req.Token,
	})
	if err != nil {
		log.Error("failed to change password", sl.Err(err))

		if status.Code(err) == codes.Unauthenticated {
			return http.StatusUnauthorized, fmt.Errorf("%s: %w", op, err)
		} else if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user password successfully changed")

	return int(resp.Status), nil
}

func (c *Client) SetPermissionLevel(ctx context.Context, req *SetPermissionLevelReq) (int, error) {
	const op = "grpc.auth.SetPermissionLevel"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.UserID),
	)

	log.Info("updating user permissions")

	if req.UserID == 0 || req.InitiatorID == 0 || req.PermissionLevel == 0 {
		log.Error("failed to change permissions", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.SetPermissionLevel(ctx, &user.SetPermissionLevelRequest{
		UserId:          req.UserID,
		InitiatorId:     req.InitiatorID,
		PermissionLevel: req.PermissionLevel,
	})
	if err != nil {
		log.Error("failed to change permissions", sl.Err(err))

		if status.Code(err) == codes.PermissionDenied {
			return http.StatusForbidden, fmt.Errorf("%s: %w", op, err)
		} else if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %s", op, "user not found or permission lvl not changed")
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user permissions updated")

	return int(resp.Status), nil
}

func (c *Client) GetPermissionLevel(ctx context.Context, userID string) (int64, error) {
	const op = "grpc.auth.GetPermissionLevel"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("UserID", userID),
	)

	log.Info("getting user permissions")

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error("failed to get permissions", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if id == 0 {
		log.Error("failed to get permissions", sl.Err(clienterrors.ErrAllFieldsRequired))

		return 0, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetPermissionLevel(ctx, &user.GetPermissionLevelRequest{
		UserId: int64(id),
	})
	if err != nil {
		log.Error("failed to get permissions", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user permissions given")

	return resp.PermissionLevel, nil
}

func (c *Client) FillUserProfile(ctx context.Context, req *FillUserProfileReq) (int, error) {
	const op = "grpc.user.FillUserProfile"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.UserID),
	)

	log.Info("filling user profile")

	if req.UserID == 0 || req.Name == "" || req.Lastname == "" || req.MiddleName == "" || req.Classname == "" || req.DateOfBirth == "" {
		log.Error("failed to fill profile", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	if req.UserID != ctx.Value("user_id") {
		log.Error("failed to fill profile", sl.Err(errors.New("access denied")))

		return http.StatusForbidden, fmt.Errorf("%s: %w", op, errors.New("access denied"))
	}

	resp, err := c.api.FillUserProfile(ctx, &user.FillUserProfileRequest{
		UserId:      req.UserID,
		Name:        req.Name,
		Lastname:    req.Lastname,
		Middlename:  req.MiddleName,
		Classname:   req.Classname,
		DateOfBirth: req.DateOfBirth,
	})
	if err != nil {
		log.Error("failed to fill profile", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("profile filled")

	return int(resp.Status), nil
}

func (c *Client) ChangeUserStatus(ctx context.Context, req *ChangeUserStatusReq) (int, error) {
	const op = "grpc.user.ChangeUserStatus"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.UserID),
	)

	log.Info("changing user status")

	if req.UserID == 0 || req.InitiatorID == 0 {
		log.Error("failed to change status", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.ChangeUserStatus(ctx, &user.ChangeUserStatusRequest{
		UserId:      req.UserID,
		InitiatorId: req.InitiatorID,
		Active:      req.Active,
	})
	if err != nil {
		log.Error("failed to change status", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user status changed")

	return int(resp.Status), nil
}

func (c *Client) IsUserActive(ctx context.Context, req *IsUserActiveReq) (bool, int, error) {
	const op = "grpc.user.IsUserActive"

	log := c.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", req.UserID),
	)

	log.Info("checking is user active")

	if req.UserID == 0 {
		log.Error("failed to check user activity", sl.Err(clienterrors.ErrAllFieldsRequired))

		return false, http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.IsUserActive(ctx, &user.IsUserActiveRequest{
		UserId: req.UserID,
	})
	if err != nil {
		log.Error("failed to check user activity", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return false, http.StatusNotFound, fmt.Errorf("%s: %w", op, clienterrors.ErrUserNotFound)
		}
		return false, http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully checked is user active")

	return resp.Active, http.StatusOK, nil
}

func (c *Client) GetStudentsByClassname(ctx context.Context, classname string) ([]*user.Student, int, error) {
	const op = "grpc.user.GetStudentsByClassname"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("Classname", classname),
	)

	log.Info("getting students by classname")

	if classname == "" {
		log.Error("failed to get students", sl.Err(clienterrors.ErrAllFieldsRequired))

		return nil, http.StatusBadRequest, fmt.Errorf("%s: %s", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.GetStudentsByClassname(ctx, &user.GetStudentsByClassnameRequest{
		Classname: classname,
	})
	if err != nil {
		log.Error("failed to get students", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return nil, http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("students are given")

	return resp.Students, http.StatusOK, nil
}

func (c *Client) DeleteUser(ctx context.Context, userID string) (int, error) {
	const op = "grpc.user.DeleteUser"

	log := c.log.With(
		slog.String("Operation", op),
		slog.String("UserID", userID),
	)

	log.Info("deleting user")

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error("failed to delete user", sl.Err(err))

		return http.StatusInternalServerError, err
	}

	if id == 0 {
		log.Error("failed to delete user", sl.Err(clienterrors.ErrAllFieldsRequired))

		return http.StatusBadRequest, fmt.Errorf("%s: %w", op, clienterrors.ErrAllFieldsRequired)
	}

	resp, err := c.api.DeleteUser(ctx, &user.DeleteUserRequest{
		UserId: int64(id),
	})
	if err != nil {
		log.Error("failed to delete user", sl.Err(err))

		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s: %w", op, err)
		}
		return http.StatusInternalServerError, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user is deleted")

	return int(resp.Status), nil
}
