package mirror

import "github.com/datacommand2/cdm-disaster-recovery/common/constant"

var mirrorEnvironmentOperationCodes = []interface{}{
	constant.MirrorEnvironmentOperationStart,
	constant.MirrorEnvironmentOperationStop,
}

var mirrorEnvironmentStateCodes = []interface{}{
	constant.MirrorEnvironmentStateCodeMirroring,
	constant.MirrorEnvironmentStateCodeStopping,
	constant.MirrorEnvironmentStateCodeError,
}

var mirrorVolumeStateCodes = []interface{}{
	constant.MirrorVolumeStateCodeWaiting,
	constant.MirrorVolumeStateCodeInitializing,
	constant.MirrorVolumeStateCodeMirroring,
	constant.MirrorVolumeStateCodePaused,
	constant.MirrorVolumeStateCodeStopping,
	constant.MirrorVolumeStateCodeDeleting,
	constant.MirrorVolumeStateCodeError,
	constant.MirrorVolumeStateCodeStopped,
}

var mirrorVolumeOperations = []interface{}{
	constant.MirrorVolumeOperationStart,
	constant.MirrorVolumeOperationStop,
	constant.MirrorVolumeOperationStopAndDelete,
	constant.MirrorVolumeOperationPause,
	constant.MirrorVolumeOperationResume,
	constant.MirrorVolumeOperationDestroy,
}

func isMirrorEnvironmentOperation(s string) bool {
	for _, code := range mirrorEnvironmentOperationCodes {
		if code == s {
			return true
		}
	}

	return false
}

func isMirrorEnvironmentStateCode(s string) bool {
	for _, code := range mirrorEnvironmentStateCodes {
		if code == s {
			return true
		}
	}

	return false
}

func isMirrorVolumeOperation(s string) bool {
	for _, code := range mirrorVolumeOperations {
		if code == s {
			return true
		}
	}
	return false
}

func isMirrorVolumeStateCode(s string) bool {
	for _, code := range mirrorVolumeStateCodes {
		if code == s {
			return true
		}
	}

	return false
}
