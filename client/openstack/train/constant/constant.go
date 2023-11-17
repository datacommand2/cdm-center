package constant

const (
	// VolumeGroupCreationCheckTimeInterval 볼륨 그룹 생성 async 처리에서 볼륨 그룹의 status 를 가져오는 호출 주기 (s)
	VolumeGroupCreationCheckTimeInterval = 1

	// VolumeGroupCreationCheckTimeOut 볼륨 그룹 생성 async 처리에 대한 timeout (s)
	VolumeGroupCreationCheckTimeOut = 120

	// VolumeGroupDeletionCheckTimeInterval 볼륨 그룹 삭제 async 처리에서 볼륨 그룹의 status 를 가져오는 호출 주기 (s)
	VolumeGroupDeletionCheckTimeInterval = 3 // 임의: 15초

	// VolumeGroupDeletionCheckTimeOut 볼륨 그룹 삭제 async 처리에 대한 timeout (s)
	VolumeGroupDeletionCheckTimeOut = 3600 // 임의: 1시간

	// VolumeGroupUpdateCheckTimeInterval 볼륨 그룹 수정 async 처리에서 볼륨 그룹의 status 를 가져오는 호출 주기 (s)
	VolumeGroupUpdateCheckTimeInterval = 1

	// VolumeGroupUpdateCheckTimeOut 볼륨 그룹 수정 async 처리에 대한 timeout (s)
	VolumeGroupUpdateCheckTimeOut = 120

	// VolumeGroupSnapshotDeletionCheckTimeInterval 볼륨 그룹 스냅샷 삭제 async 처리에서 볼륨 그룹 스냅샷의 status 를 가져오는 호출 주기 (s)
	VolumeGroupSnapshotDeletionCheckTimeInterval = 1

	// VolumeGroupSnapshotDeletionCheckTimeOut 볼륨 그룹 스냅샷 삭제 async 처리에 대한 timeout (s)
	VolumeGroupSnapshotDeletionCheckTimeOut = 120

	// VolumeGroupSnapshotCreationCheckTimeInterval 볼륨 그룹 스냅샷 생성 async 처리에서 볼륨 그룹 스냅샷의 status 를 가져오는 호출 주기 (s)
	VolumeGroupSnapshotCreationCheckTimeInterval = 3 // 임의: 15초

	// VolumeGroupSnapshotCreationCheckTimeOut 볼륨 그룹 스냅샷 생성 async 처리에 대한 timeout (s)
	VolumeGroupSnapshotCreationCheckTimeOut = 3600 // 임의: 1시간

	// VolumeCreationCheckTimeInterval 볼륨 생성 async 처리에서 볼륨의 status 를 가져오는 호출 주기 (s)
	VolumeCreationCheckTimeInterval = 1 // 임의: 1초

	// VolumeCreationCheckTimeOut 볼륨 생성 async 처리에 대한 timeout (s)
	VolumeCreationCheckTimeOut = 3600 // 임의: 1시간

	// VolumeReadOnlyCheckTimeInterval 볼륨 readonly 설정 async 처리에서 볼륨의 readonly flag 를 가져오는 호출 주기 (s)
	VolumeReadOnlyCheckTimeInterval = 1

	// VolumeReadOnlyCheckTimeOut 볼륨 readonly 설정 async 처리에 대한 timeout (s)
	VolumeReadOnlyCheckTimeOut = 30

	// VolumeDeletionCheckTimeInterval 볼륨 삭제 async 처리에서 볼륨의 status 를 가져오는 호출 주기 (s)
	VolumeDeletionCheckTimeInterval = 1

	// VolumeDeletionCheckTimeOut 볼륨 삭제 async 처리에 대한 timeout (s)
	VolumeDeletionCheckTimeOut = 120

	// VolumeUnmanagingCheckTimeInterval 볼륨 unmanaging async 처리에서 볼륨의 status 를 가져오는 호출 주기 (s)
	VolumeUnmanagingCheckTimeInterval = 1

	// VolumeUnmanagingCheckTimeOut 볼륨 unmanaging async 처리에 대한 timeout (s)
	VolumeUnmanagingCheckTimeOut = 120

	// VolumeSnapshotCreationCheckTimeInterval 볼륨 스냅샷 생성 async 처리에서 볼륨 스냅샷의 status 를 가져오는 호출 주기 (s)
	VolumeSnapshotCreationCheckTimeInterval = 1

	// VolumeSnapshotCreationCheckTimeOut 볼륨 스냅샷 생성 async 처리에 대한 timeout (s)
	VolumeSnapshotCreationCheckTimeOut = 120

	// VolumeSnapshotDeletionCheckTimeInterval 볼륨 스냅샷 삭제 async 처리에서 볼륨 스냅샷의 status 를 가져오는 호출 주기 (s)
	VolumeSnapshotDeletionCheckTimeInterval = 1

	// VolumeSnapshotDeletionCheckTimeOut 볼륨 스냅샷 삭제 async 처리에 대한 timeout (s)
	VolumeSnapshotDeletionCheckTimeOut = 120

	// VolumeSnapshotUnmanagingCheckTimeInterval 볼륨 스냅샷 unmanaging async 처리에서 볼륨 스냅샷의 status 를 가져오는 호출 주기 (s)
	VolumeSnapshotUnmanagingCheckTimeInterval = 1

	// VolumeSnapshotUnmanagingCheckTimeOut 볼륨 스냅샷 unmanaging async 처리에 대한 timeout (s)
	VolumeSnapshotUnmanagingCheckTimeOut = 120

	// InstanceCreationCheckTimeInterval 인스턴스 생성 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	InstanceCreationCheckTimeInterval = 1 // 임의: 1초

	// InstanceCreationCheckTimeOut 인스턴스 생성 async 처리에 대한 timeout (s)
	InstanceCreationCheckTimeOut = 3600 // 임의: 1시간

	// AddSecurityGroupToServerCheckTimeInterval 인스턴스에 보안그룹 추가 async 처리에서 인스턴스의 보안그룹 값을 가져오는 호출 주기 (s)
	AddSecurityGroupToServerCheckTimeInterval = 1

	// AddSecurityGroupToServerCheckTimeOut 인스턴스에 보안그룹 추가 async 처리에 대한 timeout (s)
	AddSecurityGroupToServerCheckTimeOut = 30

	// InstanceDeletionCheckTimeInterval 인스턴스 삭제 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	InstanceDeletionCheckTimeInterval = 1

	// InstanceDeletionCheckTimeOut 인스턴스 삭제 async 처리에 대한 timeout (s)
	InstanceDeletionCheckTimeOut = 120

	// InstanceForceDeletionCheckTimeInterval 인스턴스 강제 삭제 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	InstanceForceDeletionCheckTimeInterval = 1

	// InstanceForceDeletionCheckTimeOut 인스턴스 강제 삭제 async 처리에 대한 timeout (s)
	InstanceForceDeletionCheckTimeOut = 120

	// FlavorDeletionCheckTimeInterval flavor 삭제 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	FlavorDeletionCheckTimeInterval = 1

	// FlavorDeletionCheckTimeOut flavor 삭제 async 처리에 대한 timeout (s)
	FlavorDeletionCheckTimeOut = 120

	// KeypairDeletionCheckTimeInterval keypair 삭제 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	KeypairDeletionCheckTimeInterval = 1

	// KeypairDeletionCheckTimeOut keypair 삭제 async 처리에 대한 timeout (s)
	KeypairDeletionCheckTimeOut = 120

	// SourceTypeVolume 인스턴스 생성 시 사용되는 블록 디바이스의 source type ("volume" 만 사용)
	SourceTypeVolume = "volume"

	// DestinationTypeVolume 인스턴스 생성 시 사용되는 블록 디바이스의 destination type ("volume" 만 사용)
	DestinationTypeVolume = "volume"

	// InstanceStartedCheckTimeInterval 인스턴스 기동 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	InstanceStartedCheckTimeInterval = 1

	// InstanceStartedCheckTimeOut 인스턴스 기동 async 처리에 대한 timeout (s)
	InstanceStartedCheckTimeOut = 120

	// InstanceStoppedCheckTimeInterval 인스턴스 중지 async 처리에서 인스턴스의 status 를 가져오는 호출 주기 (s)
	InstanceStoppedCheckTimeInterval = 1

	// InstanceStoppedCheckTimeOut 인스턴스 중지 async 처리에 대한 timeout (s)
	InstanceStoppedCheckTimeOut = 120
)
