package config

import (
	"github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
)

// GlobalConfig
const (
	// GlobalLogLevel 로그 레벨 기본값
	GlobalLogLevel = "global_log_level"

	// GlobalLogStorePeriod 로그 보유기간 기본값
	GlobalLogStorePeriod = "global_log_store_period"

	// SystemMonitorInterval 모니터링 간격 기본값
	SystemMonitorInterval = "system_monitor_interval"

	// BackupScheduleEnable 스케쥴에 의한 솔루션 백업의 여부
	BackupScheduleEnable = "backup_schedule_enable"

	// BackupSchedule 솔루션 백업의 스케쥴
	BackupSchedule = "backup_schedule"

	// BackupStorePeriod 솔루션 백업본의 보유기간
	BackupStorePeriod = "backup_store_period"

	// BugReportEnable DataCommand 기술지원의 이메일로 버그리포트 송신 여부
	BugReportEnable = "bug_report_enable"

	// BugReportSMTP DataCommand 기술지원의 이메일로 버그리포트 송신하기 위한 SMTP 서버
	BugReportSMTP = "bug_report_smtp"

	// BugReportEmail DataCommand 기술지원의 이메일
	BugReportEmail = "bug_report_email"
)

// ServiceConfig name
const (
	// ServiceNode 서비스 노드
	ServiceNode = "service_node"

	ServiceClusterManager = "cdm-cluster-manager"

	ServiceDRManager = "cdm-dr-manager"

	ServiceMigrator = "cdm-dr-migratord"

	ServiceMirror = "cdm-dr-mirrord"

	ServiceSnapshot = "cdm-target-snapshot"

	ServiceOpenStack = "cdm-openstack-request"
)

// ServiceConfig key
const (
	// ServiceNodeType 서비스 로그 타입
	ServiceNodeType = "service_log_type"

	// ServiceLogLevel 서비스 로그 레벨
	ServiceLogLevel = "service_log_level"

	// ServiceLogStorePeriod 서비스 로그 보유기간
	ServiceLogStorePeriod = "service_log_store_period"
)

// ServiceConfig value
const (
	// ServiceNodeTypeSingle 서비스 노드 타입 싱글
	ServiceNodeTypeSingle = "single"

	// ServiceNodeTypeMultiple 서비스 노드 타입 멀티플
	ServiceNodeTypeMultiple = "multiple"
)

// TenantConfig
const (
	// GlobalTimeZone 기본 타임존
	GlobalTimeZone = "global_timezone"

	// GlobalLanguageSet 기본 언어셋
	GlobalLanguageSet = "global_language_set"

	// UserLoginRestrictionEnable 잘못된 비밀번호로 로그인 시 로그인 제한 여부
	UserLoginRestrictionEnable = "user_login_restriction_enable"

	// UserLoginRestrictionTryCount 잘못된 비밀 번호로 로그인 시 로그인 허용 횟수
	UserLoginRestrictionTryCount = "user_login_restriction_try_count"

	// UserLoginRestrictionTime 잘못된 비밀 번호로 로그인시 허용 횟수 초과후 로그인 제한 시간
	UserLoginRestrictionTime = "user_login_restriction_time"

	// UserReuseOldPassword 이전 비밀번호 설정 가능여부
	UserReuseOldPassword = "user_reuse_old_password"

	// UserPasswordChangeCycle 사용지 계정의 비밀번호 변경 주기
	UserPasswordChangeCycle = "user_password_change_cycle"

	// UserSessionTimeout 사용자 조작없이 세션을 유지하는 시간
	UserSessionTimeout = "user_session_timeout"

	// EventNotificationEnable 이벤트 알림 여부
	EventNotificationEnable = "event_notification_enable"

	// EventEmailNotificationEnable SMTP 서버를 통한 이벤트 이메일 알림 여부
	EventEmailNotificationEnable = "event_email_notification_enable"

	// EventDesktopNotificationEnable HTML5 기능을 이용한 이벤트 Desktop Notification 여부
	EventDesktopNotificationEnable = "event_desktop_notification_enable"

	// EventPopupNotificationEnable 이벤트 브라우저 팝업 알림 여부
	EventPopupNotificationEnable = "event_popup_notification_enable"

	// EventSmsNotificationEnable SMS Provider 를 통한 이벤트 SMS 알림 여부
	EventSmsNotificationEnable = "event_sms_notification_enable"

	// EventCustomNotificationEnable 이벤트 Custom 알림 여부
	EventCustomNotificationEnable = "event_custom_notification_enable"

	// EventStorePeriod 이벤트 히스토리의 보유기간
	EventStorePeriod = "event_store_period"

	// EventEmailNotifier 이메일 알림을 위한 SMTP 서버
	EventEmailNotifier = "event_smtp_notifier"

	// EventSMSNotifier SMS 알림을 위한 SMS Provider
	EventSMSNotifier = "event_sms_notifier"
)

// Config 는 Config Property 의 구조체이다.
type Config struct {
	Key   string
	Value Value
}

// GlobalConfig 는 전역 설정 값을 조회하는 함수이다.
func GlobalConfig(db *gorm.DB, key string) *Config {
	var cfg model.GlobalConfig

	if err := db.Where("key = ?", key).First(&cfg).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Warnf("Could not find global config. cause: %v", err)
		}
		return nil
	}

	return &Config{Key: key, Value: Value(cfg.Value)}
}

// TenantConfig 는 테넌트 설정 값을 조회하는 함수이다.
func TenantConfig(db *gorm.DB, tenantID uint64, key string) *Config {
	var cfg model.TenantConfig

	if err := db.Where("tenant_id = ? AND key = ?", tenantID, key).First(&cfg).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Warnf("Could not find tenant config. cause: %v", err)
		}
		return nil
	}

	return &Config{Key: key, Value: Value(cfg.Value)}
}

// ServiceConfig 는 서비스 설정 값을 조회하는 함수이다.
func ServiceConfig(db *gorm.DB, serviceName, key string) *Config {
	var cfg model.ServiceConfig

	if err := db.Where("name = ? AND key = ?", serviceName, key).First(&cfg).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Warnf("Could not find service config. cause: %v", err)
		}
		return nil
	}

	return &Config{Key: key, Value: Value(cfg.Value)}
}
