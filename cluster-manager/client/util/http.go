package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"io/ioutil"
	"net/http"
	"time"
)

// HTTPRequest http request 정보를 담는 구조체
type HTTPRequest struct {
	Header http.Header
	Method string
	URL    string
	Body   []byte
}

// HTTPResponse http response 정보를 담는 구조체
type HTTPResponse struct {
	Header http.Header
	Body   string
	Code   int
}

// Request http request 함수
func (r *HTTPRequest) Request() (*HTTPResponse, error) {
	var err error
	var req *http.Request
	var resp *http.Response

	cli := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	switch {
	case r.Method == http.MethodPost || r.Method == http.MethodPut:
		req, err = http.NewRequestWithContext(ctx, r.Method, r.URL, bytes.NewReader(r.Body))
		if err != nil {
			return nil, client.BadRequest(err)
		}
	case r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete:
		req, err = http.NewRequestWithContext(ctx, r.Method, r.URL, nil)
		if err != nil {
			return nil, client.BadRequest(err)
		}
	}

	req.Header = r.Header

	resp, err = cli.Do(req)
	if err != nil {
		return nil, client.RemoteServerError(err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			logger.Warnf("could not close response body. Cause: %+v", err)
		}
	}()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, client.RemoteServerError(err)
	}

	return &HTTPResponse{
		Header: resp.Header,
		Body:   string(respBody),
		Code:   resp.StatusCode,
	}, nil
}

func (r *HTTPResponse) Error() string {
	b, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}

	return string(b)
}
