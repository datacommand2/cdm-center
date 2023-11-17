package internal

import "github.com/datacommand2/cdm-center/cluster-manager/constant"

var (
	// OpenstackCredentialMethodPassword 비밀번호 인증
	OpenstackCredentialMethodPassword = "password"
)

// OpenstackCredentialMethods 오픈스택 인증 method
var OpenstackCredentialMethods = []interface{}{
	OpenstackCredentialMethodPassword,
}

// ClusterTypeCodes 클러스터 종류 코드
var ClusterTypeCodes = []interface{}{
	constant.ClusterTypeOpenstack,
	constant.ClusterTypeOpenshift,
	constant.ClusterTypeKubernetes,
	constant.ClusterTypeVMWare,
}

// IsClusterTypeCode 클러스터 종류 코드 여부 확인
func IsClusterTypeCode(code string) bool {
	for _, c := range ClusterTypeCodes {
		if c == code {
			return true
		}
	}
	return false
}

// ClusterPermissionModeCodes 클러스터 그룹 권한 코드
var ClusterPermissionModeCodes = []interface{}{
	constant.ClusterPermissionModeReadOnly,
	constant.ClusterPermissionModeReadWrite,
}

// IsClusterPermissionModeCode 클러스터 그룹 권한 코드 여부 확인
func IsClusterPermissionModeCode(code string) bool {
	for _, c := range ClusterPermissionModeCodes {
		if c == code {
			return true
		}
	}
	return false
}

// ClusterHypervisorTypeCodes 클러스터 하이퍼바이저 종류 코드
var ClusterHypervisorTypeCodes = []interface{}{
	constant.OpenstackHypervisorTypeKVM,
	constant.OpenstackHypervisorTypeLXC,
	constant.OpenstackHypervisorTypeQemu,
	constant.OpenstackHypervisorTypeUML,
	constant.OpenstackHypervisorTypeVmware,
	constant.OpenstackHypervisorTypeXen,
	constant.OpenstackHypervisorTypeXenserver,
	constant.OpenstackHypervisorTypeHyperv,
	constant.OpenstackHypervisorTypeVirtuozzo,
}

// IsClusterHypervisorTypeCode 클러스터 노드 종류 코드 여부 확인
func IsClusterHypervisorTypeCode(code string) bool {
	for _, c := range ClusterHypervisorTypeCodes {
		if c == code {
			return true
		}
	}
	return false
}

// OpenstackKeypairTypeCodes openstack keypair 의 타입 코드
var OpenstackKeypairTypeCodes = []interface{}{
	constant.OpenstackKeypairTypeSSH,
	constant.OpenstackKeypairTypeX509,
}

// IsOpenstackKeypairTypeCode openstack keypair 타입 코드 확인
func IsOpenstackKeypairTypeCode(code string) bool {
	for _, c := range OpenstackKeypairTypeCodes {
		if c == code {
			return true
		}
	}
	return false
}
