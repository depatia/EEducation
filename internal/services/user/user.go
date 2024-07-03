package user

import (
	"AuthService/internal/models"
	"AuthService/internal/pb"
	"AuthService/internal/services/auth"
	serviceerrors "AuthService/internal/services/service_errors"
	"AuthService/internal/utils"
	"AuthService/pkg/tools/logger/sl"
	"context"
	"log/slog"
)

type UserStore struct {
	log              *slog.Logger
	userFiller       UserFiller
	userHelper       UserHelper
	permissionGetter auth.PermissionGetter
}

func New(log *slog.Logger, userFiller UserFiller, userHelper UserHelper, permissionGetter auth.PermissionGetter) *UserStore {
	return &UserStore{
		log:              log,
		userFiller:       userFiller,
		userHelper:       userHelper,
		permissionGetter: permissionGetter,
	}
}

type UserFiller interface {
	FillUserInfo(ctx context.Context, user models.UserInfo) error
}

type UserHelper interface {
	ChangeStatus(ctx context.Context, userID int64, isActive bool) error
	IsActive(ctx context.Context, userID int64) (bool, error)
	GetStudentsByClass(ctx context.Context, classname string) ([]*models.UserDTO, error)
	DelUser(ctx context.Context, userID int64) error
}

func (s *UserStore) FillUserProfile(ctx context.Context, name, lastname, middlename, dateOfBirth, classname string, userID int64) error {
	const op = "user.FillUserProfile"

	log := s.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("filling user profile")

	user := models.UserInfo{
		ID:          userID,
		Name:        name,
		Lastname:    lastname,
		Middlename:  middlename,
		DateOfBirth: dateOfBirth,
		Classname:   classname,
	}
	if err := s.userFiller.FillUserInfo(ctx, user); err != nil {
		log.Error("failed to fill profile", sl.Err(err))

		return err
	}

	log.Info("profile filled")

	return nil
}

func (s *UserStore) ChangeUserStatus(ctx context.Context, userID int64, isActive bool, initiatorID int64) error {
	const op = "user.ChangeUserStatus"

	log := s.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("changing user status")

	lvl, err := s.permissionGetter.GetPermission(ctx, initiatorID)
	if err != nil {
		log.Error("failed to get user permissions", sl.Err(err))

		return err
	}
	if lvl < 3 {
		log.Error("failed to change status", sl.Err(serviceerrors.ErrAccessDenied))

		return serviceerrors.ErrAccessDenied
	}

	if err = s.userHelper.ChangeStatus(ctx, userID, isActive); err != nil {
		log.Error("failed to change status", sl.Err(err))

		return err
	}

	log.Info("user status changed")

	return nil
}

func (s *UserStore) IsUserActive(ctx context.Context, userID int64) (bool, error) {
	const op = "user.IsUserActive"

	log := s.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("checking is user active")

	active, err := s.userHelper.IsActive(ctx, userID)
	if err != nil {
		log.Error("failed to check user activity", sl.Err(err))

		return false, err
	}

	log.Info("successfully checked is user active")

	return active, nil
}

func (s *UserStore) GetStudentsByClassname(ctx context.Context, classname string) ([]*pb.Student, error) {
	const op = "user.GetStudentsByClassname"

	log := s.log.With(
		slog.String("Operation", op),
		slog.String("Classname", classname),
	)

	log.Info("getting students by classname")

	students, err := s.userHelper.GetStudentsByClass(ctx, classname)
	if err != nil {
		log.Error("failed to get students", sl.Err(err))

		return nil, err
	}
	if len(students) == 0 {
		log.Error("failed to get students", sl.Err(serviceerrors.ErrClassNotFound))

		return nil, serviceerrors.ErrClassNotFound
	}
	return utils.ConvertUsers(students), nil
}

func (s *UserStore) DeleteUser(ctx context.Context, userID int64) error {
	const op = "user.DeleteUser"

	log := s.log.With(
		slog.String("Operation", op),
		slog.Int64("UserID", userID),
	)

	log.Info("deleting user")

	err := s.userHelper.DelUser(ctx, userID)
	if err != nil {
		log.Error("failed to delete user", sl.Err(err))

		return err
	}

	log.Info("user is deleted")

	return nil
}
