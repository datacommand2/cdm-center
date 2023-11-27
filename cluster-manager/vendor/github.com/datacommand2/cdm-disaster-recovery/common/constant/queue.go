package constant

const (
	// QueueTriggerRecoveryJob 작업 스캐줄 실행을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		protection_group_id: <uint64>,
	//		recovery_job_id: <uint64>
	//	 }
	QueueTriggerRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job"

	// QueueTriggerWaitingRecoveryJob 작업을 실행하는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerWaitingRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-waiting"

	// QueueTriggerRunRecoveryJob 작업이 Running 상태에서 migrator 로 전달하는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerRunRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-run"

	// QueueTriggerDoneRecoveryJob 작업이 Running 상태이고 migrator 에서 작업이 완료되었음을 전달하는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerDoneRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-done"

	// QueueTriggerCancelRecoveryJob 작업이 취소 되었음을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerCancelRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-cancel"

	// QueueTriggerClearingRecoveryJob 작업이 정리 중 인걸 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerClearingRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-clearing"

	// QueueTriggerClearFailedRecoveryJob 작업이 정리 실패 되었음을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerClearFailedRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-clearing-failed"

	// QueueTriggerCompletedRecoveryJob 작업이 완료 되었음을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerCompletedRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-completed"

	// QueueTriggerReportingRecoveryJob 작업이 보고서 작성을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerReportingRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-reporting"

	// QueueTriggerFinishedRecoveryJob 작업이 종료 되었음을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		job: <migrator.RecoveryJob>
	//	 }
	QueueTriggerFinishedRecoveryJob = "cdm.disaster-recovery.trigger.recovery-job-finished"

	// QueueTriggerRecoveryJobDeleted 작업이 삭제되었음을 알리는 Queue
	// message:
	//   header: {}
	//   body: {
	//		protection_group_id: <uint64>,
	//		recovery_job_id: <uint64>
	//	 }
	QueueTriggerRecoveryJobDeleted = "cdm.disaster-recovery.trigger.recovery-job-deleted"

	// QueueAddProtectionGroupSnapshot 보호그룹 스냅샷 생성을 알리는 Queue
	// message:
	//   header: {}
	//   body: <ProtectionGroupID, uint64>
	QueueAddProtectionGroupSnapshot = "cdm.disaster-recovery.trigger.add-protection-group-snapshot"

	// QueueDeleteProtectionGroupSnapshot 보호그룹 스냅샷 삭제를 알리는 Queue
	// message:
	//   header: {}
	//   body: <ProtectionGroupID, uint64>
	QueueDeleteProtectionGroupSnapshot = "cdm.disaster-recovery.trigger.delete-protection-group-snapshot"

	// QueueDeletePlanSnapshot 보호그룹의 복구계획 스냅샷 삭제를 알리는 Queue
	// message:
	//   header: {}
	//   body: <drModel.Plan>
	QueueDeletePlanSnapshot = "cdm.disaster-recovery.trigger.delete-plan-snapshot"

	// QueueSyncPlanListByProtectionGroup 보호그룹 ID로 전체 복구계획 동기화하는 Queue
	// message:
	//   header: {}
	//   body: <ProtectionGroupID, uint64>
	QueueSyncPlanListByProtectionGroup = "cdm.disaster-recovery.trigger.sync-plan-list-by-protection-group"

	// QueueSyncAllPlansByRecoveryPlan 복구계획으로 관련 복구계획 전체 동기화하는 Queue
	// message:
	//   header: {}
	//   body: {
	//		protection_group_id: <uint64>,
	//		recovery_plan_id: <uint64>
	//	 }
	QueueSyncAllPlansByRecoveryPlan = "cdm.disaster-recovery.trigger.sync-all-plans-by-recovery-plan"

	// QueueRecoveryJobMonitor 재해복구작업 관련 모든 상태 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJob>,
	//	 }
	QueueRecoveryJobMonitor = "cdm.disaster-recovery.monitor.recovery-job"

	// QueueRecoveryJobTaskMonitor 재해복구작업 Task 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobTask>,
	//	 }
	QueueRecoveryJobTaskMonitor = "cdm.disaster-recovery.monitor.recovery-job.task"

	// QueueRecoveryJobClearTaskMonitor 재해복구작업 Clear Task 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobTask>,
	//	 }
	QueueRecoveryJobClearTaskMonitor = "cdm.disaster-recovery.monitor.recovery-job.clear-task"

	// QueueRecoveryJobTaskStatusMonitor 재해복구작업 Task 관련 모든 상태 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobTaskStatusMessage>,
	//	 }
	QueueRecoveryJobTaskStatusMonitor = "cdm.disaster-recovery.monitor.recovery-job.task.status"

	// QueueRecoveryJobTaskResultMonitor 재해복구작업 Task 관련 모든 결과 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobTaskResultMessage>,
	//	 }
	QueueRecoveryJobTaskResultMonitor = "cdm.disaster-recovery.monitor.recovery-job.task.result"

	// QueueRecoveryJobVolumeMonitor 재해복구작업 볼륨 관련 모든 상태 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobVolumeMessage>,
	//	 }
	QueueRecoveryJobVolumeMonitor = "cdm.disaster-recovery.monitor.recovery-job.volume"

	// QueueRecoveryJobInstanceMonitor 재해복구작업 인스턴스 관련 모든 상태 데이터를 주고받는 Queue
	// message:
	//   header: {}
	//   body: {
	//		<internal.RecoveryJobInstanceMessage>,
	//	 }
	QueueRecoveryJobInstanceMonitor = "cdm.disaster-recovery.monitor.recovery-job.instance"
)
