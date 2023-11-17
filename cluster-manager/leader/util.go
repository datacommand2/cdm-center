package leader

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"time"
)

func getClusterList() ([]*model.Cluster, error) {
	var clusters []*model.Cluster
	if err := database.Transaction(func(db *gorm.DB) error {
		return db.Find(&clusters).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return clusters, nil
}

func nextMidnightTicker() *time.Ticker {
	now := time.Now()

	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	midnight = midnight.Add(24 * time.Hour)

	du := midnight.Sub(now)

	return time.NewTicker(du)
}
