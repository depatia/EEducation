package grpc

import (
	"AuthService/internal/pb"
	serviceerrors "AuthService/internal/services/service_errors"
	"AuthService/internal/storage/storage"
	"AuthService/pkg/tools/jwt"
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthRepo interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, error)
	RegisterUser(
		ctx context.Context,
		email string,
		pass string,
	) (int64, error)
	Validate(
		ctx context.Context,
		token string,
	) (int64, error)
	ChangePassword(
		ctx context.Context,
		email,
		oldPassword,
		newPassword,
		token string,
	) error
	SetPermissionLevel(
		ctx context.Context,
		userID,
		permissionLevel,
		initiatorID int64,
	) error
	GetPermissionLevel(
		ctx context.Context,
		userID int64,
	) (int64, error)
}

type UserRepo interface {
	FillUserProfile(
		ctx context.Context,
		name,
		lastname,
		middlename,
		dateOfBirth,
		classname string,
		userID int64,
	) error
	ChangeUserStatus(
		ctx context.Context,
		userID int64,
		isActive bool,
		initiatorID int64,
	) error
	IsUserActive(
		ctx context.Context,
		userID int64,
	) (bool, error)
	GetStudentsByClassname(
		ctx context.Context,
		classname string,
	) ([]*pb.Student, error)
	DeleteUser(
		ctx context.Context,
		userID int64,
	) error
}

type api struct {
	pb.UnimplementedUserServiceServer
	authRepo AuthRepo
	userRepo UserRepo
}

func Register(gRPCServer *grpc.Server, authRepo AuthRepo, userRepo UserRepo) {
	pb.RegisterUserServiceServer(gRPCServer, &api{authRepo: authRepo, userRepo: userRepo})
}

func (a *api) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	token, err := a.authRepo.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "incorrect email or password")
		} else if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user with this email not found")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (a *api) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uID, err := a.authRepo.RegisterUser(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		if errors.Is(err, serviceerrors.ErrBadEmailFormat) {
			return nil, status.Error(codes.InvalidArgument, "bad email format")
		}

		return nil, status.Error(codes.Internal, "failed to register")
	}

	return &pb.RegisterResponse{
		UserId: uID,
	}, nil
}

func (a *api) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	uID, err := a.authRepo.Validate(ctx, req.Token)

	if err != nil {
		if errors.Is(err, jwt.ErrBadJWT) {
			return nil, status.Error(codes.Unauthenticated, "invalid JWT")
		}

		return nil, status.Error(codes.Internal, "failed to validate JWT")
	}

	return &pb.ValidateResponse{
		UserId: uID,
	}, nil
}

func (a *api) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	if req.NewPassword == "" {
		return &pb.UpdatePasswordResponse{Status: http.StatusBadRequest}, status.Error(codes.InvalidArgument, "new password is required")
	}

	if req.OldPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is required")
	}

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	err := a.authRepo.ChangePassword(ctx, req.Email, req.OldPassword, req.NewPassword, req.Token)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "old password is incorrect")
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to change the password")
	}

	return &pb.UpdatePasswordResponse{
		Status: http.StatusOK,
	}, nil
}

func (a *api) SetPermissionLevel(ctx context.Context, req *pb.SetPermissionLevelRequest) (*pb.SetPermissionLevelResponse, error) {
	if req.PermissionLevel == 0 {
		return nil, status.Error(codes.InvalidArgument, "permission level is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if req.InitiatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "initiator id is required")
	}

	err := a.authRepo.SetPermissionLevel(ctx, req.UserId, req.PermissionLevel, req.InitiatorId)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrAccessDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to set permissions")
	}

	return &pb.SetPermissionLevelResponse{
		Status: http.StatusOK,
	}, nil
}

func (a *api) GetPermissionLevel(ctx context.Context, req *pb.GetPermissionLevelRequest) (*pb.GetPermissionLevelResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	permLvl, err := a.authRepo.GetPermissionLevel(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get permissions")
	}

	return &pb.GetPermissionLevelResponse{
		PermissionLevel: permLvl,
	}, nil
}

func (a *api) FillUserProfile(ctx context.Context, req *pb.FillUserProfileRequest) (*pb.FillUserProfileResponse, error) {
	if req.Classname == "" {
		return nil, status.Error(codes.InvalidArgument, "classname is required")
	}

	if req.DateOfBirth == "" {
		return nil, status.Error(codes.InvalidArgument, "date of birth is required")
	}

	if req.Middlename == "" {
		return nil, status.Error(codes.InvalidArgument, "middlename is required")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Lastname == "" {
		return nil, status.Error(codes.InvalidArgument, "lastname is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := a.userRepo.FillUserProfile(ctx, req.Name, req.Lastname, req.Middlename, req.DateOfBirth, req.Classname, req.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to fill user profile")
	}

	return &pb.FillUserProfileResponse{
		Status: http.StatusOK,
	}, nil
}

func (a *api) ChangeUserStatus(ctx context.Context, req *pb.ChangeUserStatusRequest) (*pb.ChangeUserStatusResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if req.InitiatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "initiator id is required")
	}

	err := a.userRepo.ChangeUserStatus(ctx, req.UserId, req.Active, req.InitiatorId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to change user status")
	}

	return &pb.ChangeUserStatusResponse{
		Status: http.StatusOK,
	}, nil
}

func (a *api) IsUserActive(ctx context.Context, req *pb.IsUserActiveRequest) (*pb.IsUserActiveResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	active, err := a.userRepo.IsUserActive(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to check is user active")
	}

	return &pb.IsUserActiveResponse{
		Active: active,
	}, nil
}

func (a *api) GetStudentsByClassname(ctx context.Context, req *pb.GetStudentsByClassnameRequest) (*pb.GetStudentsByClassnameResponse, error) {
	if req.Classname == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	students, err := a.userRepo.GetStudentsByClassname(ctx, req.Classname)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrClassNotFound) {
			return nil, status.Error(codes.NotFound, "class not found")
		}

		return nil, status.Error(codes.Internal, "failed to check is user active")
	}

	return &pb.GetStudentsByClassnameResponse{
		Students: students,
	}, nil
}

func (a *api) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := a.userRepo.DeleteUser(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.DeleteUserResponse{
		Status: http.StatusOK,
	}, nil
}
