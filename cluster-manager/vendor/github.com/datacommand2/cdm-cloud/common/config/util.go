package config

import (
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/constant"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"
)

func getServiceLogPeriodMap() (map[string]int, error) {
	var serviceNameList = []string{
		constant.ServiceAPIGateway,
		constant.ServiceIdentity,
		constant.ServiceLicense,
		constant.ServiceNotification,
		constant.ServiceScheduler,
		ServiceClusterManager,
		ServiceDRManager,
		ServiceMigrator,
		ServiceMirror,
		ServiceSnapshot,
		ServiceOpenStack,
	}
	serviceMap := make(map[string]int)

	if err := database.Execute(func(db *gorm.DB) error {
		for _, s := range serviceNameList {
			period := 30
			cfg := ServiceConfig(db, s, ServiceLogStorePeriod)
			if cfg != nil && cfg.Value.String() != "" {
				period, _ = strconv.Atoi(cfg.Value.String())
			}

			if s == constant.ServiceAPIGateway {
				s = "cdm-cloud-api"
			}

			serviceMap[s] = period
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return serviceMap, nil
}

func DeleteExpiredServiceLogFiles() error {
	// log file 경로
	dirPath := "/var/log"
	// 폴더 내 파일 list
	// TODO: 모듈 업데이트 이후에는 ioutil.ReadDir -> os.ReadDir 수정필요
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logger.Errorf("[DeleteExpiredServiceLogFiles] Could not get log file list. Cause: %+v", err)
		return err
	}

	serviceMap, err := getServiceLogPeriodMap()
	if err != nil {
		logger.Errorf("[DeleteExpiredServiceLogFiles] Could not get log store period. Cause: %+v", err)
		return err
	}

	// 2023-01-01 형식의 compile
	cDate, _ := regexp.Compile("20+([0-9]+)-+([0-9]+)-+([0-9]+)")
	// cdm-___-___ 형식의 compile
	cService, _ := regexp.Compile("cdm-+([a-z]+)-+([a-z]+)")
	now := time.Now()

	// file.Name() ex: cdm-cluster-manager-75dcbdd7d5-qt8tr-2023-06-29.log
	for _, file := range files {
		// 파일 이름내의 서비스이름 추출
		serviceName := cService.FindAllString(file.Name(), 1)
		if len(serviceName) == 0 {
			continue
		}

		if _, ok := serviceMap[serviceName[0]]; !ok {
			continue
		}

		// 파일 이름내의 날짜형식 추출
		fileDate := cDate.FindAllString(file.Name(), 1)
		if len(fileDate) == 0 {
			continue
		}
		fileDateCvt, _ := time.Parse("2006-01-02", fileDate[0])

		// 유지된 기한 = 현재 - 파일날짜 계산
		days := int(now.Sub(fileDateCvt).Hours() / 24)

		// 유지기한 지남
		if days > serviceMap[serviceName[0]] {
			logger.Infof("[DeleteExpiredServiceLogFiles] Log file(%s) is deleted by passed(%d days) store period(%d days).",
				file.Name(), days, serviceMap[serviceName[0]])
			// 유지기한이 지난 파일은 삭제
			if err = os.Remove(fmt.Sprintf("%s/%s", dirPath, file.Name())); err != nil {
				logger.Warnf("[DeleteExpiredServiceLogFiles] Could not delete the log file(%s). Cause: %+v", file.Name(), err)
			}
		}
	}

	logger.Infof("[DeleteExpiredServiceLogFiles] Completed the deleting expired log files.")
	return nil
}
