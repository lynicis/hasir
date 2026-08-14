package csrf

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
)

type csrfInterceptor struct {
	allowedOrigins map[string]bool
}

func NewCsrfInterceptor(origins ...string) *csrfInterceptor {
	allowed := make(map[string]bool)
	for _, o := range origins {
		if o != "" {
			allowed[strings.TrimRight(o, "/")] = true
		}
	}
	return &csrfInterceptor{allowedOrigins: allowed}
}

func (i *csrfInterceptor) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip for internal server-to-server or non-HTTP calls
			if req.Spec().IsClient {
				return next(ctx, req)
			}

			origin := req.Header().Get("Origin")
			// If no Origin header, it's not a browser request (e.g., CLI, SDK). Bypass CSRF check.
			if origin == "" {
				return next(ctx, req)
			}

			// If it is a browser request, verify it's from an allowed origin
			if !i.allowedOrigins[origin] {
				return nil, connect.NewError(
					connect.CodePermissionDenied,
					errors.New("cross-origin request not allowed"),
				)
			}

			// Require the X-CSRF-Token custom header
			csrfToken := req.Header().Get("X-CSRF-Token")
			if csrfToken == "" {
				return nil, connect.NewError(
					connect.CodePermissionDenied,
					errors.New("missing X-CSRF-Token header"),
				)
			}

			return next(ctx, req)
		}
	}
}
