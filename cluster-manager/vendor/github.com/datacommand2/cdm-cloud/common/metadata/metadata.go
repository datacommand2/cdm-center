package metadata

import (
	"context"
	"encoding/json"
	"errors"
	identity "github.com/datacommand2/cdm-cloud/services/identity/proto"
	"github.com/google/uuid"
	meta "github.com/micro/go-micro/v2/metadata"
	"strconv"
	"strings"
)

// ErrNotFound 원하는 메타데이터 정보 없을 때 반환하는 에러이다.
var ErrNotFound = errors.New("not found metadata")

// context 에 세팅되는 헤더이다.
const (
	HeaderAuthenticatedSession = "X-Authenticated-Session"
	HeaderAuthenticatedUser    = "X-Authenticated-User"
	HeaderClientIP             = "X-Client-Ip"
	HeaderTenantID             = "X-Tenant-Id"
	HeaderRequestID            = "X-Request-Id"
)

// GetAuthenticatedSession 세션 정보를 가져온다.
func GetAuthenticatedSession(ctx context.Context) (string, error) {
	v, ok := meta.Get(ctx, HeaderAuthenticatedSession)
	if !ok {
		return "", ErrNotFound
	}

	return v, nil
}

// GetAuthenticatedUser 사용자 정보를 가져온다.
func GetAuthenticatedUser(ctx context.Context) (*identity.User, error) {
	v, ok := meta.Get(ctx, HeaderAuthenticatedUser)
	if !ok {
		return nil, ErrNotFound
	}

	var user identity.User
	if err := json.Unmarshal([]byte(v), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetClientIP IP 정보를 가져온다.
func GetClientIP(ctx context.Context) (string, error) {
	v, ok := meta.Get(ctx, HeaderClientIP)
	if !ok {
		return "", ErrNotFound
	}

	return v, nil
}

// GetTenantID 테넌트 ID 정보를 가져온다.
func GetTenantID(ctx context.Context) (uint64, error) {
	v, ok := meta.Get(ctx, HeaderTenantID)
	if !ok {
		return 0, ErrNotFound
	}

	return strconv.ParseUint(v, 10, 64)
}

func GetRequestID(ctx context.Context) (string, error) {
	v, ok := meta.Get(ctx, HeaderRequestID)
	if !ok {
		return "", ErrNotFound
	}

	return v, nil
}

func GenRequestID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)[0:12]
}
