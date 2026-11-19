package user

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir/api/pkg/authentication"
	"hasir/api/pkg/pagination"
	"hasir/proto/gen/go/shared"
	userv1 "hasir/proto/gen/go/user/v1"
	"hasir/proto/gen/go/user/v1/userv1connect"
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

func (h *handler) RenewTokens(
	ctx context.Context,
	req *connect.Request[userv1.RenewTokensRequest],
) (*connect.Response[userv1.RenewTokensResponse], error) {
	tokens, err := h.userService.RenewTokens(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(tokens), nil
}

func (h *handler) ForgotPassword(
	ctx context.Context,
	req *connect.Request[userv1.ForgotPasswordRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.userService.ForgotPassword(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) ResetPassword(
	ctx context.Context,
	req *connect.Request[userv1.ResetPasswordRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.userService.ResetPassword(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
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
	req *connect.Request[emptypb.Empty],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.userRepository.DeleteUser(ctx, userID); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) CreateApiKey(
	ctx context.Context,
	req *connect.Request[userv1.CreateApiKeyRequest],
) (*connect.Response[userv1.CreateApiKeyResponse], error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	apiKey := uuid.NewString()
	if err = h.userRepository.CreateApiKey(ctx, userId, req.Msg.GetName(), apiKey); err != nil {
		return nil, err
	}

	return connect.NewResponse(&userv1.CreateApiKeyResponse{
		Key: apiKey,
	}), nil
}

func (h *handler) GetApiKeys(
	ctx context.Context,
	req *connect.Request[shared.Pagination],
) (*connect.Response[userv1.KeyResponse], error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	page, pageSize := pagination.Normalize(req.Msg)

	totalCount, err := h.userRepository.GetApiKeysCount(ctx, userId)
	if err != nil {
		return nil, err
	}

	apiKeys, err := h.userRepository.GetApiKeys(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}

	var keys []*userv1.Key
	for _, apiKey := range *apiKeys {
		keys = append(keys, &userv1.Key{
			Id:   apiKey.Id,
			Name: apiKey.Name,
		})
	}

	nextPage, totalPages, err := pagination.ComputePageMeta(totalCount, page, pageSize)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&userv1.KeyResponse{
		Keys:      keys,
		NextPage:  nextPage,
		TotalPage: totalPages,
	}), nil
}

func (h *handler) CreateSshKey(
	ctx context.Context,
	req *connect.Request[userv1.CreateSshKeyRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	publicKey := strings.TrimSpace(req.Msg.GetPublicKey())
	parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid SSH public key format"))
	}

	normalizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(parsedKey)))
	if err = h.userRepository.CreateSshKey(ctx, userId, req.Msg.GetName(), normalizedKey); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetSshKeys(
	ctx context.Context,
	req *connect.Request[shared.Pagination],
) (*connect.Response[userv1.KeyResponse], error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	page, pageSize := pagination.Normalize(req.Msg)

	totalCount, err := h.userRepository.GetSshKeysCount(ctx, userId)
	if err != nil {
		return nil, err
	}

	sshKeys, err := h.userRepository.GetSshKeys(ctx, userId, page, pageSize)
	if err != nil {
		return nil, err
	}

	var keys []*userv1.Key
	for _, sshKey := range *sshKeys {
		keys = append(keys, &userv1.Key{
			Id:   sshKey.Id,
			Name: sshKey.Name,
		})
	}

	nextPage, totalPages, err := pagination.ComputePageMeta(totalCount, page, pageSize)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&userv1.KeyResponse{
		Keys:      keys,
		NextPage:  nextPage,
		TotalPage: totalPages,
	}), nil
}

func (h *handler) RevokeApiKey(
	ctx context.Context,
	req *connect.Request[userv1.RevokeKeyRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.userRepository.RevokeApiKey(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) RevokeSshKey(
	ctx context.Context,
	req *connect.Request[userv1.RevokeKeyRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.userRepository.RevokeSshKey(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}
