package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/micro/go-micro/v2/logger"
	"net/http"
	"time"
)

// ListComputeServicesRequest 컴퓨트 서비스 목록 조회 request
type ListComputeServicesRequest struct {
	Binary   string
	HostName string
}

// ListComputeServicesResult 컴퓨트 서비스 목록 조회 result
type ListComputeServicesResult struct {
	ID        string `json:"id,omitempty"`
	Binary    string `json:"binary,omitempty"`
	ZoneName  string `json:"zone,omitempty"`
	Host      string `json:"host,omitempty"`
	Status    string `json:"status,omitempty"`
	State     string `json:"state,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ListComputeServicesResponse 컴퓨트 서비스 목록 조회 response
type ListComputeServicesResponse struct {
	ComputeServices []ListComputeServicesResult ` json:"services,omitempty"`
}

// ShowHypervisorDetailsRequest 하이퍼 바이저 상세 조회 request
type ShowHypervisorDetailsRequest struct {
	HypervisorID string
}

// Service 하이퍼 바이저 서비스 구조체
type Service struct {
	Host string `json:"host,omitempty"`
}

// Server 하이퍼 바이저 내 서버 구조체
type Server struct {
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
}

// HypervisorResult 하이퍼 바이저 result
type HypervisorResult struct {
	UUID               string   `json:"id,omitempty"`
	TypeCode           string   `json:"hypervisor_type,omitempty"`
	Hostname           string   `json:"hypervisor_hostname,omitempty"`
	IPAddress          string   `json:"host_ip,omitempty"`
	VcpuTotalCnt       uint32   `json:"vcpus,omitempty"`
	VcpuUsedCnt        uint32   `json:"vcpus_used,omitempty"`
	MemTotalMibiBytes  uint64   `json:"memory_mb,omitempty"`
	MemUsedMibiBytes   uint64   `json:"memory_mb_used,omitempty"`
	DiskTotalGibiBytes uint64   `json:"local_gb,omitempty"`
	DiskUsedGibiBytes  uint64   `json:"local_gb_used,omitempty"`
	State              string   `json:"state,omitempty"`
	Status             string   `json:"status,omitempty"`
	Service            Service  `json:"service,omitempty"`
	Servers            []Server `json:"servers,omitempty"`
}

// ListHypervisorsResponse 하이퍼 바이저 목록 조회 response 구조체
type ListHypervisorsResponse struct {
	Hypervisors []HypervisorResult `json:"hypervisors,omitempty"`
}

// ShowHypervisorDetailsResponse 하이퍼 바이저 상세 조회 response 구조체
type ShowHypervisorDetailsResponse struct {
	Hypervisor HypervisorResult `json:"hypervisor,omitempty"`
	Raw        string
}

// ListHypervisors 하이퍼바이저 목록 조회
func ListHypervisors(c client.Client) (*ListHypervisorsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/os-hypervisors?with_servers=true",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListHypervisors", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListHypervisorsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowHypervisorDetails 하이퍼바이저 상세 조회
func ShowHypervisorDetails(c client.Client, req ShowHypervisorDetailsRequest) (*ShowHypervisorDetailsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/os-hypervisors/" + req.HypervisorID,
	}

	response, err := request.Request()
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ShowHypervisorDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	rsp.Raw = response.Body

	return &rsp, nil
}

// ListComputeServices 컴퓨트 서비스 목록 조회
func ListComputeServices(c client.Client, req ListComputeServicesRequest) (*ListComputeServicesResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/os-services" + "?binary=" + req.Binary + "&host=" + req.HostName,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListComputeServices", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListComputeServicesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	if len(rsp.ComputeServices) == 0 {
		return nil, client.Unknown(response)

	} else if len(rsp.ComputeServices) > 1 {
		logger.Warnf("[ListComputeServices] multiple compute services found by request")
	}

	return &rsp, nil
}

// ListAllComputeServices 컴퓨트 전체 서비스 목록 조회
func ListAllComputeServices(c client.Client) (*ListComputeServicesResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.53"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/os-services",
	}

	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllComputeServices", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListComputeServicesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
