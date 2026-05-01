package handler

import (
	"context"

	pb "github.com/jason/ecommerce/proto/user"
	"github.com/jason/ecommerce/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

func NewUserGRPCHandler(svc *service.UserService) *UserGRPCHandler {
	return &UserGRPCHandler{svc: svc}
}

func (h *UserGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.svc.Register(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register: %v", err)
	}
	return &pb.RegisterResponse{
		Id:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (h *UserGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	access, refresh, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	return &pb.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (h *UserGRPCHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return &pb.GetUserResponse{
		Id:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (h *UserGRPCHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, valid := h.svc.ValidateToken(req.Token)
	return &pb.ValidateTokenResponse{
		Valid:  valid,
		UserId: userID,
	}, nil
}
