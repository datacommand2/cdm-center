package util

import (
	"context"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/datacommand2/cdm-cloud/common/store"
	"net/http"
)

// CreateError 에러 메시지 생성
func CreateError(ctx context.Context, eventCode string, err error) error {
	if err == nil {
		return nil
	}

	var errorCode string
	switch {
	// parameter
	case errors.Equal(err, errors.ErrRequiredParameter):
		errorCode = "required_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnchangeableParameter):
		errorCode = "unchangeable_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrConflictParameterValue):
		errorCode = "conflict_parameter"
		return errors.StatusConflict(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrInvalidParameterValue):
		errorCode = "invalid_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrLengthOverflowParameterValue):
		errorCode = "length_overflow_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrOutOfRangeParameterValue):
		errorCode = "out_of_range_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnavailableParameterValue):
		errorCode = "unavailable_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrPatternMismatchParameterValue):
		errorCode = "pattern_mismatch_parameter_value"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrFormatMismatchParameterValue):
		errorCode = "format_mismatch_parameter"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	// request
	case errors.Equal(err, errors.ErrInvalidRequest):
		errorCode = "invalid_request"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnauthenticatedRequest):
		errorCode = "unauthenticated_request"
		return errors.StatusUnauthenticated(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnauthorizedRequest):
		errorCode = "unauthorized_request"
		return errors.StatusUnauthorized(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrNoContent):
		errorCode = "no_content"
		return errors.StatusNoContent(ctx, eventCode, errorCode, nil)

	// system
	case errors.Equal(err, errors.ErrUnusableDatabase):
		errorCode = "unusable_database"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnusableStore):
		errorCode = "unusable_store"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnusableBroker):
		errorCode = "unusable_broker"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, errors.ErrUnknown):
		errorCode = "unknown"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	// ipc
	case errors.Equal(err, errors.ErrIPCFailed):
		errorCode = "ipc_failed"
		if errors.GetIPCStatusCode(err) == http.StatusPreconditionFailed {
			return errors.StatusPreconditionFailed(ctx, eventCode, errorCode, err)
		}
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	// metadata
	case errors.Equal(err, metadata.ErrNotFound):
		errorCode = "not_found_tenant"
		return errors.StatusUnauthorized(ctx, eventCode, errorCode, err)

	// store
	case errors.Equal(err, store.ErrNotFoundKey):
		errorCode = "not_found_key"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	default:
		err = errors.Unknown(err)
		errorCode = "unknown"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)
	}
}
