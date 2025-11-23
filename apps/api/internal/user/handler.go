package user

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	userv1 "github.com/protohasir/proto/gen/go/user/v1"
	"github.com/protohasir/proto/gen/go/user/v1/userv1connect"
)

type handler struct {
	userv1connect.UnimplementedUserServiceHandler
	validateInterceptor connect.Interceptor
	otelInterceptor     connect.Interceptor
	userService         Service
}

func NewHandler(
	validateInterceptor connect.Interceptor,
	otelInterceptor connect.Interceptor,
	userService Service,
) *handler {
	return &handler{
		validateInterceptor: validateInterceptor,
		otelInterceptor:     otelInterceptor,
		userService:         userService,
	}
}

func (h *handler) RegisterRoutes() (string, http.Handler) {
	return userv1connect.NewUserServiceHandler(
		h,
		connect.WithInterceptors(h.validateInterceptor),
		connect.WithInterceptors(h.otelInterceptor),
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
) (*connect.Response[userv1.LoginResponse], error) {
	tokens, err := h.userService.Login(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(tokens), nil
}
