package migrator

// RecoveryJobMessage 재해복구작업의 상태값 저장하는 메시지
type RecoveryJobMessage struct {
	JobID     uint64                `json:"job_id"`
	Status    *RecoveryJobStatus    `json:"status"`
	Operation *RecoveryJobOperation `json:"operation"`
	Result    *RecoveryJobResult    `json:"result"`
	//Detail    *drms.RecoveryJob     `json:"detail"`
}

// RecoveryJobTaskStatusMessage 재해복구작업 task 상태값 저장하는 메시지
type RecoveryJobTaskStatusMessage struct {
	RecoveryJobTaskStatusKey string                 `json:"recovery_job_task_status_key,omitempty"`
	TypeCode                 string                 `json:"type_code,omitempty"`
	RecoveryJobTaskStatus    *RecoveryJobTaskStatus `json:"recovery_job_task_status"`
}

// RecoveryJobTaskResultMessage 재해복구작업 task 결과값 저장하는 메시지
type RecoveryJobTaskResultMessage struct {
	RecoveryJobTaskResultKey string                 `json:"recovery_job_task_result_key,omitempty"`
	TypeCode                 string                 `json:"type_code,omitempty"`
	IsDeleted                bool                   `json:"is_deleted,omitempty"`
	RecoveryJobTaskResult    *RecoveryJobTaskResult `json:"recovery_job_task_result"`
}

// RecoveryJobVolumeMessage 재해복구작업 볼륨 상태값 저장하는 메시지
type RecoveryJobVolumeMessage struct {
	RecoveryJobID uint64                   `json:"recovery_job_id"`
	VolumeID      uint64                   `json:"volume_id"`
	VolumeStatus  *RecoveryJobVolumeStatus `json:"volume_status"`
}

// RecoveryJobInstanceMessage 재해복구작업 인스턴스 상태값 저장하는 메시지
type RecoveryJobInstanceMessage struct {
	RecoveryJobID  uint64                     `json:"recovery_job_id"`
	InstanceID     uint64                     `json:"instance_id"`
	InstanceStatus *RecoveryJobInstanceStatus `json:"instance_status"`
}
