package constant

// topics
const (
	// TopicNoticeLicenseUpdated 라이선스 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: {}
	TopicNoticeLicenseUpdated = "cdm.cloud.notice.license.updated"

	// TopicNoticeGlobalLoggingLevelUpdated 전역 로깅 레벨 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: {}
	TopicNoticeGlobalLoggingLevelUpdated = "cdm.cloud.notice.global.logging.level.updated"

	// TopicNoticeServiceLoggingLevelUpdated 서비스 로깅 레벨 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <service-name>
	TopicNoticeServiceLoggingLevelUpdated = "cdm.cloud.notice.service.logging.level.updated"

	// TopicNoticeCreateSchedule 스케쥴 등록을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ScheduleID>
	TopicNoticeCreateSchedule = "cdm.cloud.notice.schedule.created"

	// TopicNoticeUpdateSchedule 스케쥴 수정을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ScheduleID>
	TopicNoticeUpdateSchedule = "cdm.cloud.notice.schedule.updated"

	// TopicNoticeDeleteSchedule 스케쥴 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ScheduleID>
	TopicNoticeDeleteSchedule = "cdm.cloud.notice.schedule.deleted"

	// TopicNotificationEventCreated 이벤트의 생성을 공지하는 Topic
	// tenant 마다 별개의 Topic 을 사용하며, %v 대신 tenant 의 ID 를 사용
	// message:
	//   header: {}
	//   body: {
	//     event: <model.Event>,
	//     event_code: <model.EventCode>,
	//     tenant: <model.Tenant>,
	//   }
	TopicNotificationEventCreated = "cdm.notification.event.created.%v"

	// TopicNoticeServiceLogNodeTypeUpdated 서비스 로그 노드 타입 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <type>
	TopicNoticeServiceLogNodeTypeUpdated = "cdm.cloud.notice.service.log.node.type.updated"

	// TopicNoticeSingleDeleteExpiredLogFiles single node 의 유지기한이 지난 로그 삭제 스케줄을 공지하는 Topic
	// message:
	//   header: {}
	//   body: {}
	TopicNoticeSingleDeleteExpiredLogFiles = "cdm.cloud.notice.schedule.single.delete-expired-log-files"

	// TopicNoticeMultipleDeleteExpiredLogFiles multiple node 의유지기한이 지난 로그 삭제 스케줄을 공지하는 Topic
	// message:
	//   header: {}
	//   body: {}
	TopicNoticeMultipleDeleteExpiredLogFiles = "cdm.cloud.notice.schedule.multiple.delete-expired-log-files"
)

// queues
const (
	// QueueReportEvent 이벤트 등록을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: <model.Event>
	QueueReportEvent = "cdm.cloud.report.event"

	// QueueNotifyEmail 이메일 알림을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: {
	//     user: <model.User>,
	//     event: <model.Event>
	//	 }
	QueueNotifyEmail = "cdm.notification.notify.email"

	// QueueNotifySMS SMS 알림을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: {
	//     user: <model.User>,
	//     event: <model.Event>
	//	 }
	QueueNotifySMS = "cdm.notification.notify.sms"

	// QueueNotifyDesktop Desktop 알림을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: {
	//     user: <model.User>,
	//     event: <model.Event>
	//	 }
	QueueNotifyDesktop = "cdm.notification.notify.desktop"

	// QueueNotifyBrowser Browser Popup 알림을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: {
	//     user: <model.User>,
	//     event: <model.Event>
	//	 }
	QueueNotifyBrowser = "cdm.notification.notify.browser"

	// QueueNotifyCustom Custom 알림을 위한 Persistent Queue
	// message:
	//   header: {}
	//   body: {
	//     user: <model.User>,
	//     event: <model.Event>
	//	 }
	QueueNotifyCustom = "cdm.notification.notify.custom"
)
