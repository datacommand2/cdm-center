package event

import (
	"encoding/json"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/constant"
	"github.com/datacommand2/cdm-cloud/common/database/model"
)

type options struct {
	time     int64
	contents interface{}
}

// Option options 값 설정
type Option func(*options)

// WithCreatedTime 이벤트 생성 시간 설정을 위한 옵션 함수 이다
func WithCreatedTime(time int64) Option {
	return func(o *options) {
		o.time = time
	}
}

// WithContents 이벤트 상세 내용 설정을 위한 옵션 함수 이다.
func WithContents(contents interface{}) Option {
	return func(o *options) {
		o.contents = contents
	}
}

// ReportEvent notification 서비스로 event를 생성 하도록 요청 한다.
// tenantID 이벤트 생성 시, 발생 한 tenant id
// code 이벤트 코드
func ReportEvent(tenantID uint64, eventCode, errorCode string, opts ...Option) error {
	var opt options
	for _, o := range opts {
		o(&opt)
	}

	event := model.Event{
		TenantID:  tenantID,
		Code:      eventCode,
		ErrorCode: &errorCode,
		CreatedAt: opt.time,
	}

	var bytes []byte
	var err error
	if opt.contents != nil {
		if bytes, err = json.Marshal(&opt.contents); err != nil {
			return err
		}
		event.Contents = string(bytes)
	}

	if bytes, err = json.Marshal(&event); err != nil {
		return err
	}

	return broker.Publish(constant.QueueReportEvent, &broker.Message{Body: bytes})
}
