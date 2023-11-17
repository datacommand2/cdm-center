package train

import (
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"net/http"
	"time"
)

// AssignRoleToGroupOnProjectRequest 역할 할당
type AssignRoleToGroupOnProjectRequest struct {
	ProjectUUID string
	UserUUID    string
	RoleUUID    string
}

// AssignRoleToUserOnProject 역할 할당
func AssignRoleToUserOnProject(c client.Client, role AssignRoleToGroupOnProjectRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeIdentity)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodPut,
		URL: fmt.Sprintf("%s/projects/%s/users/%s/roles/%s",
			endpoint,
			role.ProjectUUID,
			role.UserUUID,
			role.RoleUUID,
		),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("AssignRoleToUserOnProject", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			return client.NotFound(response)
		case response.Code == 409:
			// train 버전: 테넌트, 역할, 사용자가 겹쳐도 발생을 안함. 그냥 204가 떨어짐
			// TODO: 다른 버전에 대한 테스트가 필요함
			return client.Conflict(response)

		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}
