package main

import (
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/handler"
	"github.com/datacommand2/cdm-center/cluster-manager/leader"
	"github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/event"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"

	// notification for openstack
	_ "github.com/datacommand2/cdm-center/cluster-manager/client/openstack/notification"
)

// version 은 서비스 버전의 정보이다.
var version string

func getDefaultTenantID() (uint64, error) {
	var tenant model.Tenant
	if err := database.Transaction(func(db *gorm.DB) error {
		return db.Find(&tenant, model.Tenant{Name: "default"}).Error

	}); err == gorm.ErrRecordNotFound {
		return 0, errors.Unknown(err)

	} else if err != nil {
		return 0, errors.UnusableDatabase(err)
	}

	return tenant.ID, nil
}

func reportEvent(eventCode, errorCode string, eventContents interface{}) {
	tid, err := getDefaultTenantID()
	if err != nil {
		logger.Warnf("Could not report event. Cause: %+v", err)
		return
	}

	err = event.ReportEvent(tid, eventCode, errorCode, event.WithContents(eventContents))
	if err != nil {
		logger.Warnf("Could not report event. Cause: %+v", errors.Unknown(err))
	}
}

func main() {
	metadata := map[string]string{
		"CDM_SOLUTION_NAME":       constant.SolutionName,
		"CDM_SERVICE_DESCRIPTION": constant.ServiceDescription,
	}

	service := micro.NewService(
		micro.Name(constant.ServiceName),
		micro.Version(version),
		micro.Metadata(metadata),
	)

	logger.Infof("Creating service(%s)", constant.ServiceName)
	service.Init()

	// Registry 의 Kubernetes 기능을 go-micro server 옵션에 추가
	if err := service.Server().Init(server.Registry(registry.DefaultRegistry)); err != nil {
		logger.Fatalf("Cloud not init server options by service(%s). cause: %v", constant.ServiceName, err)
	}

	if err := service.Client().Init(client.Registry(registry.DefaultRegistry)); err != nil {
		logger.Fatalf("Cloud not init client options by service(%s). cause: %v", constant.ServiceName, err)
	}

	if err := selector.DefaultSelector.Init(selector.Registry(registry.DefaultRegistry)); err != nil {
		logger.Fatalf("Cloud not init selector options by service(%s). cause: %v", constant.ServiceName, err)
	}

	defer common.Destroy()

	h, err := handler.NewClusterManagerHandler()
	if err != nil {
		err = errors.Unknown(err)
		reportEvent("cdm-center.manager.main.failure-create_handler", "unknown", err)
		logger.Fatalf("Could not create handler. Cause: %+v", err)
	}

	err = cms.RegisterClusterManagerHandler(service.Server(), h)
	if err != nil {
		err = errors.Unknown(err)
		reportEvent("cdm-center.manager.main.failure-register_handler", "unknown", err)
		logger.Fatalf("Could not register handler. Cause: %+v", err)
	}

	cancel := h.StatusMonitor()
	defer cancel()

	leaderCancel := leader.Init()
	defer leaderCancel()

	logger.Infof("Running service(%s)", constant.ServiceName)
	if err = service.Run(); err != nil {
		err = errors.Unknown(err)
		reportEvent("cdm-center.manager.main.failure-run_service", "unknown", err)
		logger.Fatalf("Could not run service(%s). Cause: %+v", constant.ServiceName, err)
	}
}
