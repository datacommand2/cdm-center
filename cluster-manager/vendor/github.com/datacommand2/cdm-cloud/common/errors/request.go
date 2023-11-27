package errors

import (
	"context"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	identity "github.com/datacommand2/cdm-cloud/services/identity/proto"
)

var (
	// ErrInvalidRequest 유효하지 않은 요청
	ErrInvalidRequest = New("invalid request")

	// ErrUnauthenticatedRequest 인증되지 않은 요청
	ErrUnauthenticatedRequest = New("unauthenticated request")

	// ErrUnauthorizedRequest 인가되지 않은 요청
	ErrUnauthorizedRequest = New("unauthorized request")

	// ErrNoContent no content
	ErrNoContent = New("no content")
)

// InvalidRequest 유효하지 않은 요청
func InvalidRequest(ctx context.Context) error {
	ip, err := metadata.GetClientIP(ctx)
	if err != nil {
		ip = "unknown"
	}

	return Wrap(
		ErrInvalidRequest,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"client_ip": ip,
		}),
	)
}

// UnauthenticatedRequest 인증되지 않은 요청
func UnauthenticatedRequest(ctx context.Context) error {
	ip, err := metadata.GetClientIP(ctx)
	if err != nil {
		ip = "unknown"
	}

	key, err := metadata.GetAuthenticatedSession(ctx)
	if err != nil {
		key = "unknown"
	}

	return Wrap(
		ErrUnauthenticatedRequest,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"client_ip":   ip,
			"session_key": key,
		}),
	)
}

// UnauthorizedRequest 인가되지 않은 요청
func UnauthorizedRequest(ctx context.Context) error {
	ip, err := metadata.GetClientIP(ctx)
	if err != nil {
		ip = "unknown"
	}

	u, err := metadata.GetAuthenticatedUser(ctx)
	if err != nil {
		u = &identity.User{Name: "unknown", Account: "unknown"}
	}

	return Wrap(
		ErrUnauthorizedRequest,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"client_ip":    ip,
			"user_name":    u.Name,
			"user_account": u.Account,
		}),
	)
}
