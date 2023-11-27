package errors

import (
	"context"
	"encoding/json"
	"github.com/datacommand2/cdm-cloud/common/event"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/micro/go-micro/v2/errors"
	"net/http"
)

// Message CDM 의 메시지 구조체
type Message struct {
	Code     string `json:"code,omitempty"`
	Contents string `json:"contents,omitempty"`
}

func reportEvent(ctx context.Context, eventCode, errorCode string, eventContents interface{}) {
	id, err := metadata.GetTenantID(ctx)
	if err != nil {
		logger.Warnf("Could not report event. cause: %v", err)
		return
	}

	err = event.ReportEvent(id, eventCode, errorCode, event.WithContents(eventContents))
	if err != nil {
		logger.Warnf("Could not report event. cause: %v", err)
	}
}

func createEventMessage(eventCode, errorCode string, contents interface{}) string {
	code := eventCode
	if errorCode != "" {
		code += ":" + errorCode
	}
	var m = Message{Code: code}
	if contents != nil {
		b, _ := json.Marshal(contents)
		m.Contents = string(b)
	}

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// StatusOK generates 200 error.
func StatusOK(ctx context.Context, eventCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, "", eventContents)
	return nil
}

// StatusNoContent generates 204 error.
func StatusNoContent(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusNoContent)
}

// StatusBadRequest generates 400 error.
func StatusBadRequest(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusBadRequest)
}

// StatusUnauthenticated generates 401 error.
func StatusUnauthenticated(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusUnauthorized)
}

// StatusUnauthorized generates 403 error.
func StatusUnauthorized(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusForbidden)
}

// StatusNotFound generates 404 error.
func StatusNotFound(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusNotFound)
}

// StatusConflict generates 409 error.
func StatusConflict(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusConflict)
}

// StatusPreconditionFailed generates 412 error.
func StatusPreconditionFailed(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusPreconditionFailed)
}

// StatusInternalServerError generates 500 error.
func StatusInternalServerError(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusInternalServerError)
}

// StatusNotImplemented generates 501 error.
func StatusNotImplemented(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusNotImplemented)
}

// StatusBadGateway generates 502 error.
func StatusBadGateway(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusBadGateway)
}

// StatusServiceUnavailable generates 503 error.
func StatusServiceUnavailable(ctx context.Context, eventCode, errorCode string, eventContents interface{}) error {
	reportEvent(ctx, eventCode, errorCode, eventContents)
	return errors.New("cdm", createEventMessage(eventCode, errorCode, eventContents), http.StatusServiceUnavailable)
}
