package migrator

import "github.com/datacommand2/cdm-disaster-recovery/common/constant"

var recoveryJobOperations = []interface{}{
	constant.RecoveryJobOperationRun,
	constant.RecoveryJobOperationPause,
	constant.RecoveryJobOperationCancel,
	constant.RecoveryJobOperationRetry,
	constant.RecoveryJobOperationRollback,
	constant.RecoveryJobOperationRetryRollback,
	constant.RecoveryJobOperationIgnoreRollback,
	constant.RecoveryJobOperationConfirm,
	constant.RecoveryJobOperationRetryConfirm,
	constant.RecoveryJobOperationCancelConfirm,
}

func isRecoveryJobOperation(s string) bool {
	for _, op := range recoveryJobOperations {
		if op == s {
			return true
		}
	}

	return false
}

var recoveryJobStateCodes = []interface{}{
	constant.RecoveryJobStateCodeWaiting,
	constant.RecoveryJobStateCodePending,
	constant.RecoveryJobStateCodeRunning,
	constant.RecoveryJobStateCodeCanceling,
	constant.RecoveryJobStateCodePaused,
	constant.RecoveryJobStateCodeCompleted,
	constant.RecoveryJobStateCodeClearing,
	constant.RecoveryJobStateCodeClearFailed,
	constant.RecoveryJobStateCodeReporting,
	constant.RecoveryJobStateCodeFinished,
}

func isRecoveryJobStateCodes(s string) bool {
	for _, code := range recoveryJobStateCodes {
		if code == s {
			return true
		}
	}

	return false
}

var recoveryResultCodes = []interface{}{
	constant.RecoveryResultCodeSuccess,
	constant.RecoveryResultCodePartialSuccess,
	constant.RecoveryResultCodeFailed,
	constant.RecoveryResultCodeCanceled,
}

func isRecoveryResultCode(c string) bool {
	for _, code := range recoveryResultCodes {
		if code == c {
			return true
		}
	}

	return false
}
