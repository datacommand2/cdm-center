package watcher

import (
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/config"
	"github.com/datacommand2/cdm-cloud/common/constant"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	mlogger "github.com/micro/go-micro/v2/logger"
)

var defaultWatcher *watcher

type watcher struct {
	serviceName string

	gSub broker.Subscriber
	sSub broker.Subscriber
}

func (w *watcher) findServiceLoggingLevel(db *gorm.DB, s string) *mlogger.Level {
	cfg := config.ServiceConfig(db, s, config.ServiceLogLevel)
	if cfg != nil {
		lv, _ := mlogger.GetLevel(cfg.Value.String())
		return &lv
	}

	return nil
}

func (w *watcher) findGlobalLoggingLevel(db *gorm.DB) *mlogger.Level {
	cfg := config.GlobalConfig(db, config.GlobalLogLevel)
	if cfg != nil {
		lv, _ := mlogger.GetLevel(cfg.Value.String())
		return &lv
	}

	return nil
}

func (w *watcher) updateLoggingLevel() {
	var lv *mlogger.Level

	err := database.Transaction(func(db *gorm.DB) error {
		// find service logging level
		if w.serviceName != "" {
			lv = w.findServiceLoggingLevel(db, w.serviceName)
		}

		// find global logging level
		if lv == nil {
			lv = w.findGlobalLoggingLevel(db)
		}

		return nil
	})
	if err != nil {
		logger.Warnf("Could not get logging level. cause: %v", err)
	}

	// apply logging level
	if lv != nil {
		_ = logger.Init(mlogger.WithLevel(*lv))
	}
}

func (w *watcher) watchGlobalLoggingLevelUpdated() (err error) {
	w.gSub, err = broker.SubscribeTempQueue(
		constant.TopicNoticeGlobalLoggingLevelUpdated,
		func(_ broker.Event) error {
			w.updateLoggingLevel()
			return nil
		},
	)

	return
}

func (w *watcher) watchServiceLoggingLevelUpdated() (err error) {
	w.sSub, err = broker.SubscribeTempQueue(
		constant.TopicNoticeServiceLoggingLevelUpdated,
		func(e broker.Event) error {
			if w.serviceName == "" {
				return nil
			}

			if w.serviceName == string(e.Message().Body) {
				w.updateLoggingLevel()
			}

			return nil
		},
	)

	return
}

func (w *watcher) watch() error {
	// update logging level from database
	w.updateLoggingLevel()

	// subscribe global logging level updated notice
	if err := w.watchGlobalLoggingLevelUpdated(); err != nil {
		logger.Warnf("Could not subscribe global logging level updated notice. cause: %v", err)
	}

	// subscribe global logging level updated notice
	if err := w.watchServiceLoggingLevelUpdated(); err != nil {
		logger.Warnf("Could not subscribe service logging level updated notice. cause: %v", err)
	}

	return nil
}

func (w *watcher) stop() error {
	// unsubscribe global logging level updated notice
	if w.gSub != nil {
		if err := w.gSub.Unsubscribe(); err != nil {
			logger.Warnf("Could not unsubscribe global logging level updated notice. cause: %v", err)
		}
	}

	// unsubscribe service logging level updated notice
	if w.sSub != nil {
		if err := w.sSub.Unsubscribe(); err != nil {
			logger.Warnf("Could not unsubscribe service logging level updated notice. cause: %v", err)
		}
	}

	return nil
}

// Watch 는 전역 혹은 서비스 로깅레벨 변경을 감지하여 default logger 의 로깅레벨을 변경해주는 함수이다.
func Watch(serviceName string) error {
	if defaultWatcher != nil {
		if err := defaultWatcher.stop(); err != nil {
			logger.Warnf("Could not stop watcher. cause: %v", err)
		}
	}

	w := &watcher{
		serviceName: serviceName,
	}

	if err := w.watch(); err != nil {
		return err
	}

	defaultWatcher = w

	return nil
}

// Stop 은 전역 혹은 서비스 로깅레벨 변경을 감지를 중지하는 함수이다.
func Stop() error {
	if defaultWatcher == nil {
		return nil
	}

	defer func() {
		defaultWatcher = nil
	}()

	return defaultWatcher.stop()
}
