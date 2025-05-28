package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/AleksZelenchuk/vault-server/gen/go/vaultuserpb"
	"github.com/AleksZelenchuk/vault-server/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "log"
	"strconv"

	"github.com/AleksZelenchuk/vault-server/pkg/storage"
	"github.com/google/uuid"
)

type UserVaultService struct {
	vaultuserpb.UnimplementedVaultUserServiceServer
	store *storage.UserStore
	// publisher can be used for Redis PubSub broadcasting
}

func NewUserVaultService(store *storage.UserStore) *UserVaultService {
	return &UserVaultService{store: store}
}

// Register will create new user from given data
// todo: add password confirmation field to validate before save
func (s *UserVaultService) Register(ctx context.Context, req *vaultuserpb.CreateUserRequest) (*vaultuserpb.CreateUserResponse, error) {
	if req == nil || req.User == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: user data is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user := &storage.User{
		ID:       uuid.New(),
		Email:    req.User.Email,
		Username: req.User.Username,
		Password: hashedPassword,
	}
	result, err := s.store.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	lastInsertedId, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &vaultuserpb.CreateUserResponse{Id: strconv.FormatInt(lastInsertedId, 10)}, nil
}

// Login - perform user login with a given username and password, return either error or generated JWT token
func (s *UserVaultService) Login(ctx context.Context, req *vaultuserpb.LoginRequest) (*vaultuserpb.LoginResponse, error) {
	user, err := s.store.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &vaultuserpb.LoginResponse{Token: token}, nil
}

// Deprecated: GetUserByUsername not sure if this is needed
// probably will rework it to retrieve user by its active id to avoid security issues
func (s *UserVaultService) GetUserByUsername(ctx context.Context, req *vaultuserpb.GetUserRequest) (*vaultuserpb.GetUserResponse, error) {
	_, err := auth.UserIDFromContext(ctx)
	if err != true {
		return nil, errors.New("no user id provided")
	}

	user, err2 := s.store.GetByUsername(ctx, req.Username)
	if err2 != nil {
		return nil, err2
	}

	return &vaultuserpb.GetUserResponse{User: userToProto(user)}, nil
}

func (s *UserVaultService) DeleteUser(ctx context.Context, req *vaultuserpb.DeleteUserRequest) (*vaultuserpb.DeleteUserResponse, error) {
	_, err := auth.UserIDFromContext(ctx)
	if err != true {
		return nil, errors.New("no user id provided")
	}

	id, err2 := uuid.Parse(req.Id)
	if err2 != nil {
		return nil, err2
	}

	success, err2 := s.store.DeleteUser(ctx, id)
	if err2 != nil {
		return nil, err2
	}

	return &vaultuserpb.DeleteUserResponse{Success: success}, nil
}

// convert user data to proto format
func userToProto(e *storage.User) *vaultuserpb.VaultUser {
	return &vaultuserpb.VaultUser{
		Id:       e.ID.String(),
		Email:    e.Email,
		Username: e.Username,
		Password: string(e.Password),
	}
}
