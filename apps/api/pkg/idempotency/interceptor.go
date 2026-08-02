package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"connectrpc.com/connect"
	"github.com/goccy/go-json"
)

const HeaderKey = "Idempotency-Key"

type IdempotencyRepository interface {
	IsIdempotencyKeyExists(ctx context.Context, key string) (error, bool)
	ClaimIdempotencyKey(ctx context.Context, key, operation, requestHash string) error
	EndOperation(ctx context.Context, key, response, err string) error
}

type idempotencyInterceptor struct {
	repo IdempotencyRepository
}

func NewIdempotencyInterceptor(repo IdempotencyRepository) *idempotencyInterceptor {
	return &idempotencyInterceptor{
		repo: repo,
	}
}

func (i *idempotencyInterceptor) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Spec().IsClient {
				return next(ctx, req)
			}

			if req.Spec().IdempotencyLevel != connect.IdempotencyIdempotent {
				return next(ctx, req)
			}

			idempotencyKey := req.Header().Get(HeaderKey)
			if idempotencyKey == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New("no idempotency key provided in header"),
				)
			}

			err, exists := i.repo.IsIdempotencyKeyExists(ctx, idempotencyKey)
			if err != nil {
				return nil, err
			}

			if exists {
				return nil, connect.NewError(
					connect.CodeAlreadyExists,
					errors.New("idempotency key already claimed"),
				)
			}

			operation := req.Spec().Procedure
			requestHash, err := i.hashRequestBody(req.Any())
			if err != nil {
				return nil, connect.NewError(
					connect.CodeInternal,
					errors.New("couldn't hash request body"),
				)
			}

			err = i.repo.ClaimIdempotencyKey(ctx, idempotencyKey, operation, requestHash)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeInternal,
					errors.New("couldn't claim idempotency key"),
				)
			}

			resp, nextErr := next(ctx, req)

			var hashedResponse string
			var errStr string
			if nextErr != nil {
				errStr = nextErr.Error()
			} else {
				hashedResponse, err = i.hashRequestBody(resp.Any())
				if err != nil {
					// We couldn't hash the response, but the operation succeeded.
					// We'll record a fallback response hash to allow the operation to end successfully.
					hashedResponse = "hash_failed"
				}
			}

			if err = i.repo.EndOperation(ctx, idempotencyKey, hashedResponse, errStr); err != nil {
				return nil, connect.NewError(
					connect.CodeInternal,
					errors.New("couldn't end operation"),
				)
			}

			return resp, nextErr
		}
	}
}

func (i *idempotencyInterceptor) hashRequestBody(body any) (string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
