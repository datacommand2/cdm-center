package cluster

import (
	"context"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	commonModel "github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"regexp"
	"unicode/utf8"
)

type credentialValidator interface {
	Validate(string) error
}

func newCredentialValidator(code string) (credentialValidator, error) {
	switch code {
	// TODO: 다른 클러스터도 validator 추가
	case constant.ClusterTypeOpenstack:
		return &openstackCredentialValidator{}, nil
	}

	return nil, internal.UnsupportedClusterType(code)
}

type openstackCredentialValidator struct{}

func (v *openstackCredentialValidator) Validate(credential string) error {
	if len(credential) > 1024 {
		return errors.LengthOverflowParameterValue("cluster.credential", "", 1024)
	}

	var cred internal.OpenstackCredential
	if err := json.Unmarshal([]byte(credential), &cred); err != nil {
		return errors.FormatMismatchParameterValue("cluster.credential", "", "ClusterCredential")
	}

	if cred.Methods == nil {
		return errors.RequiredParameter("cluster.credential.methods")
	}

	if len(cred.Methods) != 1 {
		return errors.OutOfRangeParameterValue("cluster.credential.methods", cred.Methods, 1, 1)
	}

	switch cred.Methods[0] {
	case internal.OpenstackCredentialMethodPassword:
		if cred.Password.User.Name == "" {
			return errors.RequiredParameter("cluster.credential.password.user.name")
		}
		if cred.Password.User.Domain.Name == "" {
			return errors.RequiredParameter("cluster.credential.password.user.domain.name")
		}
		if cred.Password.User.Password == "" {
			return errors.RequiredParameter("cluster.credential.password.user.password")
		}

	default:
		return errors.UnavailableParameterValue("cluster.credential.methods", "", internal.OpenstackCredentialMethods)
	}

	return nil
}

// CheckValidClusterConnectionInfo 클러스터 접속정보 유효성 확인
func CheckValidClusterConnectionInfo(conn *cms.ClusterConnectionInfo) error {
	apiServerURL := conn.GetApiServerUrl()
	credential := conn.GetCredential()
	typeCode := conn.GetTypeCode()
	tenantID := conn.GetTenantId()

	if typeCode == "" {
		return errors.RequiredParameter("cluster.type_code")
	}

	switch typeCode {
	case constant.ClusterTypeOpenstack:
		if apiServerURL == "" {
			return errors.RequiredParameter("cluster.api_server_url")
		}
		if len(apiServerURL) > 300 {
			return errors.LengthOverflowParameterValue("cluster.api_server_url", apiServerURL, 300)
		}
		if !govalidator.IsURL(apiServerURL) {
			return errors.FormatMismatchParameterValue("cluster.api_server_url", apiServerURL, "URL")
		}

		if credential == "" {
			return errors.RequiredParameter("cluster.credential")
		}
		if len(credential) > 1024 {
			return errors.LengthOverflowParameterValue("cluster.credential", "", 1024)
		}
		if v, err := newCredentialValidator(typeCode); err != nil {
			return err
		} else if err := v.Validate(credential); err != nil {
			return err
		}

		if tenantID != "" {
			if _, err := uuid.Parse(tenantID); err != nil {
				return errors.FormatMismatchParameterValue("cluster.tenant.uuid", tenantID, "UUID")
			}
		}
	//case internal.ClusterTypeKubernetes:
	//case internal.ClusterTypeOpenshift:
	//case internal.ClusterTypeVMWare:
	default:
		return internal.UnsupportedClusterType(typeCode)
	}

	return nil
}

func validateAddClusterRequest(ctx context.Context, db *gorm.DB, req *cms.AddClusterRequest, credential string) error {
	if req.GetCluster() == nil {
		return errors.RequiredParameter("cluster")
	}

	if req.GetCluster().GetId() != 0 {
		return errors.InvalidParameterValue("cluster.id", req.Cluster.Id, "unusable parameter")
	}

	user, _ := metadata.GetAuthenticatedUser(ctx)
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, req.Cluster.OwnerGroup.Id) {
		return errors.UnauthorizedRequest(ctx)
	}
	if err := CheckValidClusterConnectionInfo(&cms.ClusterConnectionInfo{
		TypeCode:     req.Cluster.TypeCode,
		ApiServerUrl: req.Cluster.ApiServerUrl,
		Credential:   credential,
	}); err != nil {
		return err
	}

	if req.Cluster.OwnerGroup == nil || req.Cluster.OwnerGroup.Id == 0 {
		return errors.RequiredParameter("cluster.owner_group.id")

	} else if err := db.First(&commonModel.Group{}, req.Cluster.OwnerGroup.Id).Error; err == gorm.ErrRecordNotFound {
		return errors.InvalidParameterValue("cluster.owner_group.id", req.Cluster.OwnerGroup.Id, "not found group")

	} else if err != nil {
		return errors.UnusableDatabase(err)
	}

	if req.Cluster.Name == "" {
		return errors.RequiredParameter("cluster.name")
	}
	if utf8.RuneCountInString(req.Cluster.Name) > 255 {
		return errors.LengthOverflowParameterValue("cluster.name", req.Cluster.Name, 255)
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9가-힣\\.\\#\\-\\_]*$", req.Cluster.Name)
	if !matched {
		return errors.InvalidParameterValue("group.name", req.Cluster.Name, "Only Hangul, Alphabet and special characters (-, _, #, .) are allowed.")
	}

	tid, _ := metadata.GetTenantID(ctx)
	err := db.Where(&model.Cluster{TenantID: tid, Name: req.Cluster.Name}).First(&model.Cluster{}).Error
	switch {
	case err == nil:
		return errors.ConflictParameterValue("cluster.name", req.Cluster.Name)
	case err != gorm.ErrRecordNotFound:
		return errors.UnusableDatabase(err)
	}

	err = db.Where(&model.Cluster{TenantID: tid, APIServerURL: req.Cluster.ApiServerUrl}).First(&model.Cluster{}).Error
	switch {
	case err == nil:
		return errors.ConflictParameterValue("cluster.ApiServerUrl", req.Cluster.ApiServerUrl)
	case err != gorm.ErrRecordNotFound:
		return errors.UnusableDatabase(err)
	}

	if utf8.RuneCountInString(req.Cluster.Remarks) > 300 {
		return errors.LengthOverflowParameterValue("cluster.remarks", req.Cluster.Remarks, 300)
	}

	return nil
}

func validateUpdateClusterRequest(ctx context.Context, db *gorm.DB, req *cms.UpdateClusterRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetCluster() == nil {
		return errors.RequiredParameter("cluster")
	}

	if req.Cluster.OwnerGroup == nil || req.Cluster.OwnerGroup.Id == 0 {
		return errors.RequiredParameter("cluster.owner_group.id")

	} else if err := db.First(&commonModel.Group{}, req.Cluster.OwnerGroup.Id).Error; err == gorm.ErrRecordNotFound {
		return errors.InvalidParameterValue("cluster.owner_group.id", req.Cluster.OwnerGroup.Id, "not found group")

	} else if err != nil {
		return errors.UnusableDatabase(err)
	}

	if req.Cluster.Name == "" {
		return errors.RequiredParameter("cluster.name")
	}

	if utf8.RuneCountInString(req.Cluster.Name) > 255 {
		return errors.LengthOverflowParameterValue("cluster.name", req.Cluster.Name, 255)
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9가-힣\\.\\#\\-\\_]*$", req.Cluster.Name)
	if !matched {
		return errors.InvalidParameterValue("group.name", req.Cluster.Name, "Only Hangul, Alphabet, Number(0-9) and special characters (-, _, #, .) are allowed.")
	}

	tid, _ := metadata.GetTenantID(ctx)
	err := db.Not(&model.Cluster{ID: req.ClusterId}).
		Where(&model.Cluster{TenantID: tid, Name: req.Cluster.Name}).
		First(&model.Cluster{}).Error
	switch {
	case err == nil:
		return errors.ConflictParameterValue("cluster.name", req.Cluster.Name)
	case err != gorm.ErrRecordNotFound:
		return errors.UnusableDatabase(err)
	}

	if utf8.RuneCountInString(req.Cluster.Remarks) > 300 {
		return errors.LengthOverflowParameterValue("cluster.remarks", req.Cluster.Remarks, 300)
	}

	return nil
}

// 수정 요청에 대한 유효성을 검사한다. 단, credential 정보는 별도의 함수에서 수정되며, 요청 시 공백으로 전달되므로
// 별도의 검사를 하지 않는다. 데이터 저장은 이후 루틴에서 "이름"과 "기타"만 저장하므로 검사를 하지 않아도 무방하다.
func checkUpdatableCluster(orig *model.Cluster, c *cms.Cluster) error {
	if orig.ID != c.GetId() {
		return errors.UnchangeableParameter("cluster.id")
	}

	if c.OwnerGroup != nil && c.OwnerGroup.Id != 0 && c.OwnerGroup.Id != orig.OwnerGroupID {
		return errors.UnchangeableParameter("cluster.owner_group.id")
	}

	if c.TypeCode != "" && c.TypeCode != orig.TypeCode {
		return errors.UnchangeableParameter("cluster.type_code")
	}

	if c.ApiServerUrl != "" && c.ApiServerUrl != orig.APIServerURL {
		return errors.UnchangeableParameter("cluster.api_server_url")
	}

	return nil
}

func validateGetClusters(ctx context.Context, req *cms.ClusterListRequest) error {
	// validity check
	if len(req.TypeCode) != 0 && !internal.IsClusterTypeCode(req.TypeCode) {
		return errors.UnavailableParameterValue("type_code", req.TypeCode, internal.ClusterTypeCodes)
	}
	if utf8.RuneCountInString(req.Name) > 255 {
		return errors.LengthOverflowParameterValue("name", req.Name, 255)
	}

	if req.GetOwnerGroupId() != 0 {
		user, _ := metadata.GetAuthenticatedUser(ctx)
		if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, req.GetOwnerGroupId()) {
			return errors.UnauthorizedRequest(ctx)
		}
	}

	return nil
}
