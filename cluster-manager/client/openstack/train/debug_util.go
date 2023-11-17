package train

import (
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"os"
	"sync"
	"time"
)

var requestLogLock = sync.Mutex{}

func authRequestLog(funcName string, requestTime time.Time, request util.HTTPRequest, response *util.HTTPResponse, resErr error) {
	s := time.Now().Local().Format("2006-01-02")
	f, err := os.OpenFile(fmt.Sprintf("/var/log/cdm-openstack-request-%v.log", s), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		logger.Infof("file open error")
	} else {
		//f.WriteString(fmt.Sprintf("%v - Reqeust[%v]\n", time.Now(), funcName))
		f.WriteString(fmt.Sprintf("%v - Reqeust[%v]\n", requestTime, funcName))
		curl := "curl -v -s -X"
		curl = curl + " " + request.Method
		curl = curl + " " + request.URL
		f.WriteString(curl + "\n")

		f.WriteString(fmt.Sprintf("%v - Response[%v]\n", time.Now(), funcName))
		if response == nil {
			f.WriteString("Response is nil\n")
		} else {
			//f.WriteString(fmt.Sprintf("Response Body : %v\n", response.Body))
			f.WriteString(fmt.Sprintf("Response Code : %v\n", response.Code))
		}

		if resErr != nil {
			f.WriteString(fmt.Sprintf("http request error : %v", err))
		}
		f.WriteString("\n")
		f.Close()
	}
}

func requestLog(funcName string, requestTime time.Time, c client.Client, request util.HTTPRequest, response *util.HTTPResponse, resErr error) {
	requestLogLock.Lock()
	defer func() {
		requestLogLock.Unlock()
	}()

	s := time.Now().Local().Format("2006-01-02")
	f, err := os.OpenFile(fmt.Sprintf("/var/log/cdm-openstack-request-%v.log", s), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		logger.Infof("file open error")
	} else {
		f.WriteString(fmt.Sprintf("%v - Reqeust[%v]\n", requestTime, funcName))
		curl := "curl -v -s -X"
		curl = curl + " " + request.Method
		curl = curl + " " + request.URL
		curl = curl + " -H \"X-Auth-Token:" + c.GetTokenKey() + "\""
		curl = curl + " -H \"accept:" + request.Header.Get("Content-Type") + "\""
		curl = curl + " -H \"OpenStack-API-Version:" + request.Header.Get("OpenStack-API-Version") + "\""
		curl = curl + " -H \"Content-Type:" + request.Header.Get("Content-Type") + "\""
		if len(string(request.Body[:])) != 0 {
			curl = curl + " -d " + string(request.Body[:])
		}

		f.WriteString(curl + "\n")

		f.WriteString(fmt.Sprintf("%v : (%v) - Response[%v]\n", time.Now(), time.Since(requestTime), funcName))
		if response == nil {
			f.WriteString("Response is nil\n")
		} else {
			f.WriteString(fmt.Sprintf("Response Body : %v\n", response.Body))
			f.WriteString(fmt.Sprintf("Response Code : %v\n", response.Code))
		}

		if resErr != nil {
			f.WriteString(fmt.Sprintf("http request error : %v", err))
		}
		f.WriteString("\n")
		f.Close()
	}
}
