package scheduler

import (
	"encoding/json"
	"github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
)

// Model scheduler.Schedule -> model.Schedule
func (x *Schedule) Model() (*model.Schedule, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.Schedule
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

// SetFromModel model.Schedule -> scheduler.Schedule
func (x *Schedule) SetFromModel(m *model.Schedule) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}
