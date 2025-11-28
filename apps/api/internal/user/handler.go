package user

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/user/v1/userv1connect"
	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

type handler struct {
	interceptors   []connect.Interceptor
	userService    Service
	userRepository Repository
}

func NewHandler(
	userService Service,
	userRepository Repository,
	interceptors ...connect.Interceptor,
) *handler {
	return &handler{
		userService:    userService,
		userRepository: userRepository,
		interceptors:   interceptors,
	}
}

func (h *handler) RegisterRoutes() (string, http.Handler) {
	return userv1connect.NewUserServiceHandler(
		h,
		connect.WithInterceptors(h.interceptors...),
	)
}

func (h *handler) Register(
	ctx context.Context,
	req *connect.Request[userv1.RegisterRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.userService.Register(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) Login(
	ctx context.Context,
	req *connect.Request[userv1.LoginRequest],
) (*connect.Response[userv1.TokenEnvelope], error) {
	tokens, err := h.userService.Login(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(tokens), nil
}

func (h *handler) UpdateUser(
	ctx context.Context,
	req *connect.Request[userv1.UpdateUserRequest],
) (*connect.Response[userv1.TokenEnvelope], error) {
	tokens, err := h.userService.UpdateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(tokens), nil
}

func (h *handler) DeleteAccount(
	ctx context.Context,
	req *connect.Request[userv1.DeleteAccountRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.userRepository.DeleteUser(ctx, req.Msg.UserId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}
