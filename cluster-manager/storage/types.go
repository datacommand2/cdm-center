package storage

const (
	// ClusterStorageTypeNFS NFS 스토리지 타입
	ClusterStorageTypeNFS = "openstack.storage.type.nfs"

	// ClusterStorageTypeCeph ceph 스토리지 타입
	ClusterStorageTypeCeph = "openstack.storage.type.ceph"

	// ClusterStorageTypeLvm lvm 스토리지 타입
	ClusterStorageTypeLvm = "openstack.storage.type.lvm"

	// ClusterStorageTypeUnKnown 알 수 없는 스토리지 타입
	ClusterStorageTypeUnKnown = "openstack.storage.type.unknown"
)

// ClusterStorageTypeCodes 클러스터 스토리지 종류
var ClusterStorageTypeCodes = []interface{}{
	ClusterStorageTypeNFS,
	ClusterStorageTypeCeph,
	ClusterStorageTypeLvm,
	ClusterStorageTypeUnKnown,
}

// IsClusterStorageTypeCode 클러스터 매니저 서비스에서 지원 하는 스토리지 타입 여부 확인 함수
func IsClusterStorageTypeCode(s string) bool {
	for _, c := range ClusterStorageTypeCodes {
		if s == c {
			return true
		}
	}

	return false
}
