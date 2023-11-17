package train

import (
	"context"
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	util2 "github.com/datacommand2/cdm-cloud/common/util"
	"github.com/micro/go-micro/v2/metadata"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// openstack service type
const (
	// ServiceTypeIdentity identity service
	ServiceTypeIdentity = "identity"
	// ServiceTypeCompute compute service
	ServiceTypeCompute = "compute"
	// ServiceTypePlacement placement service
	ServiceTypePlacement = "placement"
	// ServiceTypeNetwork network service
	ServiceTypeNetwork = "network"
	// ServiceTypeVolumeV3 volume v3 service
	ServiceTypeVolumeV3 = "volumev3"
	// ServiceTypeVolumeV2 volume v2 service
	ServiceTypeVolumeV2 = "volumev2"
	// ServiceTypeImage image service
	ServiceTypeImage = "image"
)

// header 연관 정보
const (
	// HeaderXAuthToken 토큰 정보
	HeaderXAuthToken = "X-Auth-Token"
	// HeaderXSubjectToken 토큰 정보
	HeaderXSubjectToken = "X-Subject-Token"
)

// Connection openstack 연결을 위한 구조체
type Connection struct {
	Token     *Token
	endpoints map[string]string
	Mutex     sync.Mutex
	Close     chan interface{}
}

// AuthenticationRequest openstack 로그인 request
type AuthenticationRequest struct {
	URL  string
	Auth string
}

// GetAvailableProjectScopeResult project scope result
type GetAvailableProjectScopeResult struct {
	UUID string `json:"id,omitempty"`
}

// GetAvailableProjectScopesResponse project scope 목록 조회 response 구조체
type GetAvailableProjectScopesResponse struct {
	Projects []GetAvailableProjectScopeResult `json:"projects,omitempty"`
}

// IsConnected openstack 연결 확인
func (c *Connection) IsConnected() bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	return c.Token != nil
}

// GetAvailableProjectScopes project scope 목록 조회
func GetAvailableProjectScopes(c client.Client, url string) (*GetAvailableProjectScopesResponse, error) {
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    url + "/v3/auth/projects",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("GetAvailableProjectScopes", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		}
	}

	var rsp GetAvailableProjectScopesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// Authentication openstack 로그인
func (c *Connection) Authentication(req AuthenticationRequest) error {
	logger.Debug("Connecting to Openstack.")

	header := http.Header{}
	header.Set("Content-Type", "application/json")

	request := util.HTTPRequest{
		Header: header,
		Method: http.MethodPost,
		URL:    req.URL + "/v3/auth/tokens",
		Body:   []byte(req.Auth),
	}

	//request를 통해 이전 토큰 확인
	hToken, find := VerifyToken(req)

	var response *util.HTTPResponse
	var err error

	//토큰 없을시 접속 및 갱신
	if find == false {
		t := time.Now()
		response, err = request.Request()
		authRequestLog("Authentication", t, request, response, err)
		if err != nil {
			return err
		}

		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 404:
			return client.Unknown(response)
		case response.Code == 401:
			return client.UnAuthenticated(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code >= 400:
			return client.Unknown(response)
		case response.Code >= 500:
			//TODO:
			// 1. Context 전달 받을 방법이 없어 임시로 1로 처리.
			// 2. 어디서부터 받은 리퀘스트인지 에러에 포함시키여야함.
			ctx := context.Background()
			ctx = metadata.Set(ctx, "X-Tenant-Id", strconv.Itoa(1))
			util2.CreateError(ctx, "openstack_server_error", errors.ErrIPCFailed)
			return client.RemoteServerError(response)
		}

		token := Token{Key: response.Header.Get(HeaderXSubjectToken)}
		if err := json.Unmarshal([]byte(response.Body), &token); err != nil {
			return errors.Unknown(err)
		}

		//토근 갱신
		UpdateToken(req, &token)
		hToken = &token
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Token = hToken
	c.endpoints = make(map[string]string)

	for _, catalog := range c.Token.Payload.Catalog {
		if len(catalog.Endpoints) > 0 {
			// interface 별로 서비스 엔트포인트 URL 이 조회 되는데
			//		{
			//			"endpoints": [
			//				{
			//					"id": "068d1b359ee84b438266cb736d81de97",
			//					"interface": "public",
			//					"region": "RegionOne",
			//					"region_id": "RegionOne",
			//					"url": "http://example.com/identity"
			//				},
			//				{
			//					"id": "8bfc846841ab441ca38471be6d164ced",
			//					"interface": "admin",
			//					"region": "RegionOne",
			//					"region_id": "RegionOne",
			//					"url": "http://example.com/identity"
			//				},
			//				{
			//					"id": "beb6d358c3654b4bada04d4663b640b9",
			//					"interface": "internal",
			//					"region": "RegionOne",
			//					"region_id": "RegionOne",
			//					"url": "http://example.com/identity"
			//				}
			//			],
			//			"type": "identity",
			//			"id": "050726f278654128aba89757ae25950c",
			//			"name": "keystone"
			//		}
			// interface 에 대한 연구가 되어있지 않았지만
			// interface 별로 URL 이 모두 같아 interface 와 관계없이 첫 URL 을 사용한다.

			u1, _ := url.Parse(req.URL)
			ip1, _, _ := net.SplitHostPort(u1.Host)
			u2, _ := url.Parse(catalog.Endpoints[0].URL)
			ip2, _, _ := net.SplitHostPort(u2.Host)
			c.endpoints[catalog.Type] = strings.Replace(catalog.Endpoints[0].URL, ip2, ip1, -1)
		}
	}

	return nil
}

// Reconnect openstack 연결 재시도
func (c *Connection) Reconnect(url, auth string) {
	for {
		c.Mutex.Lock()
		if c.Token == nil {
			c.Mutex.Unlock()
			return
		}

		ticker := time.NewTicker(30 * time.Second)

		c.Mutex.Unlock()

		select {
		case <-c.Close:
			ticker.Stop()
			return

		case <-ticker.C:
			if err := c.Authentication(AuthenticationRequest{URL: url, Auth: auth}); err != nil {
				if errors.Equal(err, client.ErrUnAuthenticated) || errors.Equal(err, client.ErrUnAuthorized) {
					logger.Errorf("[Reconnect] Could not reconnect to Openstack: url(%s) auth(%s). Cause: %+v", url, auth, err)
					ticker.Stop()
					return
				}
				logger.Warnf("[Reconnect] Could not reconnect to Openstack. Cause: %+v", err)
			}
			ticker.Stop()
			continue
		}
	}
}

// CheckAuthority API 요청 권한 확인
func (c *Connection) CheckAuthority() error {
	Admin := "admin"

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	//logger.Infof("[CheckAuthority] Payload %+v", c.Token.Payload)
	if c.Token.Payload.Project.Name == Admin {
		for _, r := range c.Token.Payload.Roles {
			if r.Name == Admin {
				return nil
			}
		}
	}

	return client.UnAuthorizedUser(c.Token.Payload.Project.Name, c.Token.Payload.User.Name, Admin)
}

// GetTokenKey Token 키 조회
func (c *Connection) GetTokenKey() string {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Token == nil {
		return ""
	}
	return c.Token.Key
}

// GetEndpoint 엔드포인트 조회
func (c *Connection) GetEndpoint(serviceType string) (string, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if val, ok := c.endpoints[serviceType]; ok {
		if serviceType == ServiceTypeIdentity {
			if exist := strings.Contains(val, "/v3"); !exist {
				val += "/v3"
			}
		}
		return val, nil
	}

	return "", client.NotFoundEndpoint(serviceType)
}

var clusterTokenMap = make(map[AuthenticationRequest]*Token)
var clusterTokenLock = sync.Mutex{}

// VerifyToken 토큰 조회 및 폐기
func VerifyToken(request AuthenticationRequest) (*Token, bool) {
	clusterTokenLock.Lock()
	defer clusterTokenLock.Unlock()
	if token, ok := clusterTokenMap[request]; ok {
		t := time.Now().UTC()                // 현재 시간
		verifyTime := t.Add(time.Minute * 1) // 현재 시간 + 1분
		if verifyTime.Before(token.Payload.ExpiresAt) {
			//logger.Infof("[VerifyToken] Get key %v, token %v, check %v expires %v", request, token.Payload.Project, verifyTime, token.Payload.ExpiresAt)
			return token, true
		}
		logger.Infof("[VerifyToken] Delete token %v, expires %v", request.URL, token.Payload.ExpiresAt)
		delete(clusterTokenMap, request)
	}
	return nil, false
}

// UpdateToken 토근 갱신
func UpdateToken(request AuthenticationRequest, token *Token) {
	clusterTokenLock.Lock()
	defer clusterTokenLock.Unlock()
	if token != nil {
		logger.Infof("[UpdateToken] Update key %v, expires %v", request.URL, token.Payload.ExpiresAt)
		clusterTokenMap[request] = token
	}
}

// DeleteToken 토큰 삭제. 오픈스택에서 에러 응답을 보낼경우 세션이 만료되는 경우가 있으므로 재사용을 위한 토큰 맵에서 삭제시켜야한다.
func DeleteToken(tokenKey string) {
	clusterTokenLock.Lock()
	defer clusterTokenLock.Unlock()

	for k, v := range clusterTokenMap {
		if tokenKey == v.Key {
			delete(clusterTokenMap, k)
		}
	}
}
