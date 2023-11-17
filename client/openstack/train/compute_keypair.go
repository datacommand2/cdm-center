package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net/http"
	"time"
)

// KeypairResult keypair result
type KeypairResult struct {
	Name        string `json:"name,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	TypeCode    string `json:"type,omitempty"`
}

// ShowKeypairDetailsRequest Keypair 상세 조회 request
type ShowKeypairDetailsRequest struct {
	KeypairName string
}

// ShowKeypairDetailsResponse Keypair 상세 조회 response
type ShowKeypairDetailsResponse struct {
	Keypair KeypairResult `json:"keypair,omitempty"`
}

// ListKeypairsResponse Keypair 목록 조회 response
type ListKeypairsResponse struct {
	Keypairs []ShowKeypairDetailsResponse `json:"keypairs,omitempty"`
}

// CreateKeypairRequest Keypair 생성 request
type CreateKeypairRequest struct {
	Keypair KeypairResult `json:"keypair,omitempty"`
}

// CreateKeypairResponse Keypair 생성 response
type CreateKeypairResponse struct {
	Keypair KeypairResult `json:"keypair,omitempty"`
}

// DeleteKeypairRequest Keypair 삭제 request
type DeleteKeypairRequest struct {
	KeypairName string
}

// ListKeypairs Keypair 목록 조회
func ListKeypairs(c client.Client) (*ListKeypairsResponse, error) {
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
		URL:    endpoint + "/os-keypairs",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListKeypairs", t, c, request, response, err)
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

	var rsp ListKeypairsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowKeypairDetails Keypair 상세 조회
func ShowKeypairDetails(c client.Client, req ShowKeypairDetailsRequest) (*ShowKeypairDetailsResponse, error) {
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
		URL:    endpoint + "/os-keypairs/" + req.KeypairName,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowKeypairDetails", t, c, request, response, err)
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
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ShowKeypairDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ImportOrCreateKeypair Keypair import (혹은 생성)
func ImportOrCreateKeypair(c client.Client, req CreateKeypairRequest) (*CreateKeypairResponse, error) {
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, errors.Unknown(err)
	}

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
		Method: http.MethodPost,
		URL:    endpoint + "/os-keypairs",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ImportOrCreateKeypair", t, c, request, response, err)
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
		case response.Code == 409:
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp CreateKeypairResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// keypair 삭제 완료 여부 확인
func checkKeypairDeletion(c client.Client, name string) error {
	timeout := time.After(time.Second * constant.KeypairDeletionCheckTimeOut)

	for {
		_, err := ShowKeypairDetails(c, ShowKeypairDetailsRequest{
			KeypairName: name,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil

		case err != nil:
			return err
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.KeypairDeletionCheckTimeInterval):
			continue
		}
	}
}

// DeleteKeypair Keypair 삭제
func DeleteKeypair(c client.Client, req DeleteKeypairRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodDelete,
		URL:    endpoint + "/os-keypairs/" + req.KeypairName,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteKeypair", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("keypair could not be found.")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	if err = checkKeypairDeletion(c, req.KeypairName); err != nil {
		return err
	}

	return nil
}
