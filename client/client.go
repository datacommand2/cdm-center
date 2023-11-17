package client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
)

type clusterClientCreationFunc func(string, string, string) Client

var clusterClientCreationFuncMap map[string]clusterClientCreationFunc

// Key 는 오픈스택 credential 정보를 복호화하기위한 Key 값이다.
var Key string

// RegisterClusterClientCreationFunc 는 클러스터 타입별 Client 구조체 생성 함수의 맵이다.
func RegisterClusterClientCreationFunc(typeCode string, fn clusterClientCreationFunc) {
	if clusterClientCreationFuncMap == nil {
		clusterClientCreationFuncMap = make(map[string]clusterClientCreationFunc)
	}

	clusterClientCreationFuncMap[typeCode] = fn
}

// Client 는 클러스터의 정보를 받아오는 함수를 정의한 인터페이스
type Client interface {
	Connect() error
	Close() error
	CheckAuthority() error

	// GetTokenKey 토큰 키 조회
	GetTokenKey() string
	// GetEndpoint 엔드포인트 조회
	GetEndpoint(string) (string, error)

	// CreateTenant 테넌트 생성
	CreateTenant(CreateTenantRequest) (*CreateTenantResponse, error)
	// DeleteTenant 테넌트 삭제
	DeleteTenant(DeleteTenantRequest) error
	// GetTenantList 테넌트 목록 조회
	GetTenantList() (*GetTenantListResponse, error)
	// GetTenant 테넌트 상세 조회
	GetTenant(GetTenantRequest) (*GetTenantResponse, error)

	// GetHypervisorList 하이퍼바이저 목록 조회
	GetHypervisorList() (*GetHypervisorListResponse, error)
	// GetHypervisor 하이퍼바이저 상세 조회
	GetHypervisor(GetHypervisorRequest) (*GetHypervisorResponse, error)
	// GetAvailabilityZoneList 가용구역 목록 조회
	GetAvailabilityZoneList() (*GetAvailabilityZoneListResponse, error)
	// GetInstanceList 인스턴스 목록 조회
	GetInstanceList() (*GetInstanceListResponse, error)
	// GetInstance 인스턴스 상세 조회
	GetInstance(GetInstanceRequest) (*GetInstanceResponse, error)
	// GetInstanceSpec 인스턴스 spec 상세 조회
	GetInstanceSpec(GetInstanceSpecRequest) (*GetInstanceSpecResponse, error)
	// GetInstanceExtraSpec 인스턴스 extra spec 상세 조회
	GetInstanceExtraSpec(GetInstanceExtraSpecRequest) (*GetInstanceExtraSpecResponse, error)
	// GetKeyPairList 키페어 목록 조회
	GetKeyPairList() (*GetKeyPairListResponse, error)
	// GetKeyPair 키페어 상세 조회
	GetKeyPair(GetKeyPairRequest) (*GetKeyPairResponse, error)

	// CreateSecurityGroup security group 생성
	CreateSecurityGroup(CreateSecurityGroupRequest) (*CreateSecurityGroupResponse, error)
	// CreateSecurityGroupRule security group rule 생성
	CreateSecurityGroupRule(CreateSecurityGroupRuleRequest) (*CreateSecurityGroupRuleResponse, error)
	// DeleteSecurityGroup security group 삭제
	DeleteSecurityGroup(DeleteSecurityGroupRequest) error
	// DeleteSecurityGroupRule security group rule 삭제
	DeleteSecurityGroupRule(DeleteSecurityGroupRuleRequest) error
	// GetSecurityGroupList security group 목록 조회
	GetSecurityGroupList() (*GetSecurityGroupListResponse, error)
	// GetSecurityGroup security group 상세 조회
	GetSecurityGroup(GetSecurityGroupRequest) (*GetSecurityGroupResponse, error)
	// GetSecurityGroupRule security group rule 상세 조회
	GetSecurityGroupRule(GetSecurityGroupRuleRequest) (*GetSecurityGroupRuleResponse, error)

	// CreateNetwork 네트워크 생성
	CreateNetwork(CreateNetworkRequest) (*CreateNetworkResponse, error)
	// DeleteNetwork 네트워크 삭제
	DeleteNetwork(DeleteNetworkRequest) error
	// GetNetworkList 네트워크 목록 조회
	GetNetworkList() (*GetNetworkListResponse, error)
	// GetNetwork 네트워크 상세 조회
	GetNetwork(GetNetworkRequest) (*GetNetworkResponse, error)

	// CreateSubnet 서브넷 생성
	CreateSubnet(CreateSubnetRequest) (*CreateSubnetResponse, error)
	// DeleteSubnet 서브넷 삭제
	DeleteSubnet(DeleteSubnetRequest) error
	// GetSubnet 서브넷 상세 조회
	GetSubnet(GetSubnetRequest) (*GetSubnetResponse, error)

	// CreateRouter 라우터 생성
	CreateRouter(CreateRouterRequest) (*CreateRouterResponse, error)
	// DeleteRouter 라우터 삭제
	DeleteRouter(DeleteRouterRequest) error
	// GetRouterList 라우터 목록 조회
	GetRouterList() (*GetRouterListResponse, error)
	// GetRouter 라우터 상세 조회
	GetRouter(GetRouterRequest) (*GetRouterResponse, error)

	// CreateFloatingIP 부동 IP 생성
	CreateFloatingIP(CreateFloatingIPRequest) (*CreateFloatingIPResponse, error)
	// DeleteFloatingIP 부동 IP 삭제
	DeleteFloatingIP(DeleteFloatingIPRequest) error
	// GetFloatingIP floating ip 상세 조회
	GetFloatingIP(GetFloatingIPRequest) (*GetFloatingIPResponse, error)

	// GetStorageList 스토리지 목록 조회
	GetStorageList() (*GetStorageListResponse, error)
	// GetStorage 스토리지 상세 조회
	GetStorage(GetStorageRequest) (*GetStorageResponse, error)
	// GetVolumeSnapshotList 볼륨 스냅샷 목록 조회
	GetVolumeSnapshotList() (*GetVolumeSnapshotListResponse, error)
	// GetVolumeSnapshot 볼륨 스냅샷 상세 조회
	GetVolumeSnapshot(GetVolumeSnapshotRequest) (*GetVolumeSnapshotResponse, error)
	// GetVolumeList 볼륨 목록 조회
	GetVolumeList() (*GetVolumeListResponse, error)
	// GetVolume 볼륨 상세 조회
	GetVolume(GetVolumeRequest) (*GetVolumeResponse, error)

	// CopyVolume 볼륨 copy
	CopyVolume(CopyVolumeRequest) (*CopyVolumeResponse, error)
	// DeleteVolumeCopy 볼륨 copy 삭제
	DeleteVolumeCopy(DeleteVolumeCopyRequest) error

	// GetVolumeGroupList 볼륨 그룹 목록 조회
	GetVolumeGroupList() (*GetVolumeGroupListResponse, error)
	// GetVolumeGroup 볼륨 그룹 조회
	GetVolumeGroup(GetVolumeGroupRequest) (*GetVolumeGroupResponse, error)
	// CreateVolumeGroup 볼륨 그룹 생성
	CreateVolumeGroup(CreateVolumeGroupRequest) (*CreateVolumeGroupResponse, error)
	// DeleteVolumeGroup 볼륨 그룹 삭제
	DeleteVolumeGroup(DeleteVolumeGroupRequest) error
	// UpdateVolumeGroup 볼륨 그룹 수정
	UpdateVolumeGroup(UpdateVolumeGroupRequest) error

	// CreateVolumeGroupSnapshot 볼륨 그룹 스냅샷 생성
	CreateVolumeGroupSnapshot(CreateVolumeGroupSnapshotRequest) (*CreateVolumeGroupSnapshotResponse, error)
	// DeleteVolumeGroupSnapshot 볼륨 그룹 스냅샷 삭제
	DeleteVolumeGroupSnapshot(DeleteVolumeGroupSnapshotRequest) error
	// GetVolumeGroupSnapshotList 볼륨 그룹 스냅샷 목록 조회
	GetVolumeGroupSnapshotList() (*GetVolumeGroupSnapshotListResponse, error)

	// CreateVolume 볼륨 생성
	CreateVolume(CreateVolumeRequest) (*CreateVolumeResponse, error)
	// DeleteVolume 볼륨 삭제
	DeleteVolume(DeleteVolumeRequest) error
	// UnmanageVolume 볼륨 unmanage
	UnmanageVolume(UnmanageVolumeRequest) error

	// CreateVolumeSnapshot 볼륨 스냅샷 생성
	CreateVolumeSnapshot(CreateVolumeSnapshotRequest) (*CreateVolumeSnapshotResponse, error)
	// DeleteVolumeSnapshot 볼륨 스냅샷 삭제
	DeleteVolumeSnapshot(DeleteVolumeSnapshotRequest) error
	// ImportVolume 볼륨 import
	ImportVolume(ImportVolumeRequest) (*ImportVolumeResponse, error)

	// CreateInstanceSpec 인스턴스 스팩 생성
	CreateInstanceSpec(CreateInstanceSpecRequest) (*CreateInstanceSpecResponse, error)
	// DeleteInstanceSpec 인스턴스 스팩 삭제
	DeleteInstanceSpec(DeleteInstanceSpecRequest) error
	// CreateKeypair Keypair 생성
	CreateKeypair(CreateKeypairRequest) (*CreateKeypairResponse, error)
	// DeleteKeypair Keypair 삭제
	DeleteKeypair(DeleteKeypairRequest) error
	// CreateInstance 인스턴스 생성
	CreateInstance(CreateInstanceRequest) (*CreateInstanceResponse, error)
	// DeleteInstance 인스턴스 삭제
	DeleteInstance(DeleteInstanceRequest) error
	// StartInstance 인스턴스 기동
	StartInstance(StartInstanceRequest) error
	// StopInstance 인스턴스 중지
	StopInstance(StopInstanceRequest) error
	// PreprocessCreatingInstance 인스턴스 생성 전처리
	PreprocessCreatingInstance(PreprocessCreatingInstanceRequest) error
}

// New 는 클러스터에 요청하기 위한 Client 를 만드는 함수
func New(typeCode, url, credential, tenantID string) (Client, error) {
	if fn, ok := clusterClientCreationFuncMap[typeCode]; ok {
		return fn(url, credential, tenantID), nil
	}

	return nil, internal.UnsupportedClusterType(typeCode)
}

// Connect 클러스터 접속
func Connect(typeCode, url, credential, tenantID string, fn func(Client) error) error {
	cli, err := New(typeCode, url, credential, tenantID)
	if err != nil {
		return err
	}

	if err := cli.Connect(); err != nil {
		return err
	}

	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("Could not close client. Cause: %+v", err)
		}
	}()

	return fn(cli)
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func decryptAESCBC(cipherText []byte, key []byte, iv []byte) (plainText []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	plainText = make([]byte, len(cipherText))

	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(plainText, cipherText)

	plainText = pkcs5Trimming(plainText)
	return
}

// DecryptCredentialPassword 는 복호화
func DecryptCredentialPassword(password string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", errors.Unknown(err)
	}

	plainText, err := decryptAESCBC(cipherText, []byte(Key), []byte(Key[8:24]))
	if err != nil {
		return "", errors.Unknown(err)
	}

	return string(plainText), nil
}
