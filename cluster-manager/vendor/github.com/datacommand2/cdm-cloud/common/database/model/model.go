package model

import (
	"database/sql"
	"time"
)

// Tenant 테넌트
type Tenant struct {
	// ID 는 자동 증가되는 테넌트의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// Name 은 테넌트 이름이다.
	Name string `gorm:"not null;index" json:"name,omitempty"`

	// Remarks 는 테넌트의 비고이다.
	Remarks *string `json:"remarks,omitempty"`

	// UseFlag 는 테넌트의 사용여부이다.
	UseFlag bool `gorm:"not null;index" json:"use_flag,omitempty"`

	// CreatedAt 는 테넌트이 생성된 시간(UnixTime)이다.
	CreatedAt int64 `json:"created_at,omitempty"`

	// UpdatedAt 는 테넌트 정보가 수정된 최종시간(UnixTime)이다.
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

// TenantSolution 테넌트가 사용할 수 있는 솔루션
type TenantSolution struct {
	// TenantID 는 테넌트의 ID 이다.
	TenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"tenant_id,omitempty"`

	// Solution 은 솔루션 이름이다.
	Solution string `gorm:"primary_key;autoIncrement:false" json:"solution,omitempty"`
}

// TenantReceiveEvent 테넌트의 이벤트 수신 여부
type TenantReceiveEvent struct {
	// Code 는 알림여부 대상 이벤트의 코드이다.
	Code string `gorm:"primary_key" json:"code,omitempty"`

	// TenantID 은 알림여부 대상 테넌트의 ID 이다.
	TenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"tenant_id,omitempty"`

	// ReceiveFlag 는 알림여부이다.
	ReceiveFlag bool `gorm:"not null" json:"receive_flag,omitempty"`
}

// Role 솔루션 역할
type Role struct {
	// ID 는 자동 증가되는 역할의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// Solution 은 역할의 대상 솔루션이다.
	Solution string `gorm:"not null;unique_index:cdm_role_un" json:"solution,omitempty"`

	// Role 은 역할이다.
	Role string `gorm:"not null;unique_index:cdm_role_un" json:"role,omitempty"`
}

// Group 사용자 그룹
type Group struct {
	// ID 는 자동 증가되는 그룹의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// TenantID 는 사용자 그룹가 속한 테넌트의 ID 이다.
	TenantID uint64 `gorm:"not null" json:"tenant_id,omitempty"`

	// Name 은 그룹 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Remarks 는 그룹의 비고이다.
	Remarks *string `json:"remarks,omitempty"`

	// CreatedAt 는 그룹이 생성된 시간(UnixTime)이다.
	CreatedAt int64 `json:"created_at,omitempty"`

	// UpdatedAt 는 그룹 정보가 수정된 최종시간(UnixTime)이다.
	UpdatedAt int64 `json:"updated_at,omitempty"`

	// DeletedFlag 는 그룹 삭제 시 설정되는 값이다
	DeletedFlag bool `json:"deleted_flag,omitempty"`
}

// User 사용자
type User struct {
	// ID 는 자동 증가되는 사용자의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// TenantID 는 사용자가 속한 테넌트의 ID 이다.
	TenantID uint64 `gorm:"not null" json:"tenant_id,omitempty"`

	// Account 는 사용자가 로그인하기 위해 사용하는 계정명이다.
	Account string `gorm:"not null;unique" json:"account,omitempty"`

	// Name 은 사용자 이름이다.
	Name string `gorm:"not null;index" json:"name,omitempty"`

	// Department 는 사용자의 부서명이다.
	Department *string `json:"department,omitempty"`

	// Position 은 사용자의 직책이다.
	Position *string `json:"position,omitempty"`

	// Email 은 사용자의 이메일 주소이다.
	Email *string `gorm:"unique" json:"email,omitempty"`

	// Contact 는 사용자의 연락처이다.
	Contact *string `json:"contact,omitempty"`

	// LanguageSet 은 사용자의 사용 언어셋이다.
	LanguageSet *string `json:"language_set,omitempty"`

	// Timezone 은 사용자의 타임존이다.
	Timezone *string `json:"timezone,omitempty"`

	// Password 는 SHA-256 으로 해싱된 사용자 비밀번호이다.
	Password string `gorm:"not null" json:"password,omitempty"`

	// OldPassword 는 SHA-256 으로 해싱된 사용자의 변경전 비밀번호이다.
	OldPassword *string `json:"old_password,omitempty"`

	// PasswordUpdatedAt 은 사용자의 최종 비밀번호 수정시간(UnixTime)이다.
	PasswordUpdatedAt *int64 `json:"password_updated_at,omitempty"`

	// PasswordUpdateFlag 는 사용자의 비밀번호 수정 필요여부이다.
	PasswordUpdateFlag *bool `json:"password_update_flag,omitempty"`

	// LastLoggedInAt 은 사용자가 최종 접속한 시간(UnixTime)이다.
	LastLoggedInAt *int64 `json:"last_logged_in_at,omitempty"`

	// LastLoggedInIP 는 사용자가 최종 접속한 IP 주소 이다.
	LastLoggedInIP *string `json:"last_logged_in_ip,omitempty"`

	// LastLoginFailedCount 는 로그인에 연속 실패한 횟수이다.
	// 로그인에 성공하면 초기화된다.
	LastLoginFailedCount *uint `json:"last_login_failed_count,omitempty"`

	// LastLoginFailedAt 는 로그인에 실패한 마지막 시간(UnixTime)이다.
	// 로그인에 성공하면 초기화된다.
	LastLoginFailedAt *int64 `json:"last_login_failed_at,omitempty"`

	// CreatedAt 는 사용자가 생성된 시간(UnixTime)이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// UpdatedAt 는 사용자 정보가 수정된 최종시간(UnixTime)이다.
	UpdatedAt int64 `gorm:"not null" json:"updated_at,omitempty"`
}

// UserRole 는 사용자의 솔루션 역할 목록이다.
type UserRole struct {
	// UserID 는 사용자의 ID 이다.
	UserID uint64 `gorm:"primary_key;autoIncrement:false"`

	// RoleID 는 솔루션 역할의 ID 이다.
	RoleID uint64 `gorm:"primary_key;autoIncrement:false"`
}

// UserGroup 은 사용자의 그룹 목록이다.
type UserGroup struct {
	// UserID 는 사용자의 ID 이다.
	UserID uint64 `gorm:"primary_key;autoIncrement:false"`

	// GroupID 는 그룹의 ID 이다.
	GroupID uint64 `gorm:"primary_key;autoIncrement:false"`
}

// UserReceiveEvent 사용자의 이벤트 수신 여부
type UserReceiveEvent struct {
	// Code 는 알림여부 대상 이벤트의 코드이다.
	Code string `gorm:"primary_key" json:"code,omitempty"`

	// UserID 는 알림여부 대상 사용자의 ID 이다.
	UserID uint64 `gorm:"primary_key;autoIncrement:false" json:"user_id,omitempty"`

	// ReceiveFlag 는 알림여부이다.
	ReceiveFlag sql.NullBool `json:"receive_flag,omitempty"`
}

// Event 이벤트 히스토리
type Event struct {
	// ID 는 자동 증가되는 이벤트의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// TenantID 는 이벤트가 발생한 테넌트의 ID 이다.
	TenantID uint64 `gorm:"not null;index" json:"tenant_id,omitempty"`

	// Code 는 발생한 이벤트의 코드이다.
	Code string `gorm:"not null;index" json:"code,omitempty"`

	// ErrorCode 는 발생한 이벤트의 에러이다.
	ErrorCode *string `json:"error_code,omitempty"`

	// Contents 는 발생한 이벤트에 대한 JSON 형식의 상세 내용이다.
	Contents string `gorm:"not null" json:"contents,omitempty"`

	// CreatedAt 는 이벤트가 발생한 시간(UnixTime)이다.
	CreatedAt int64 `gorm:"not null;index" json:"created_at,omitempty"`
}

// EventCode 이벤트 코드
type EventCode struct {
	// Code 는 발생한 이벤트의 코드이며 이벤트 코드 식별자이다.
	Code string `gorm:"primary_key" json:"code,omitempty"`

	// Solution 은 솔루션 이름이다.
	Solution string `gorm:"not null;index" json:"solution,omitempty"`

	// Level 은 이벤트의 경고 수준이며, critical, warn, info, trace 중 하나만 허용한다.
	Level string `gorm:"not null;index" json:"level,omitempty"`

	// Class1 은 이벤트의 대분류이다.
	Class1 string `gorm:"column:class_1;not null;index:cdm_event_code_class" json:"class_1,omitempty"`

	// Class2 은 이벤트의 중분류이다.
	Class2 string `gorm:"column:class_2;not null;index:cdm_event_code_class" json:"class_2,omitempty"`

	// Class3 은 이벤트의 소분류이다.
	Class3 string `gorm:"column:class_3;not null;index:cdm_event_code_class" json:"class_3,omitempty"`
}

// EventCodeMessage 이벤트 코드 메시지
type EventCodeMessage struct {
	// Code 는 발생한 이벤트의 코드이다.
	Code string `gorm:"primary_key;" json:"code,omitempty"`

	// Language 은 detail 의 언어이며 eng,kor 식으로 입력한다.
	Language string `gorm:"not null;index" json:"language,omitempty"`

	// Brief 은 이벤트에 대한 요약이다.
	Brief string `gorm:"not null;index" json:"brief,omitempty"`

	// Detail 은 이벤트에 대한 설명이다.
	Detail string `gorm:"not null;index" json:"detail,omitempty"`
}

// EventError 이벤트 코드 에러
type EventError struct {
	// Code 는 발생한 에러의 코드이다.
	Code string `gorm:"primary_key;" json:"code,omitempty"`

	// Solution 은 솔루션 이름이다.
	Solution string `gorm:"not null;index" json:"solution,omitempty"`

	// Service 는 서비스 이름이다.
	Service string `gorm:"not null;index" json:"service,omitempty"`

	// Contents 는 에러 내용이다.
	Contents string `gorm:"not null;index" json:"contents,omitempty"`
}

// Schedule 스케쥴
type Schedule struct {
	// ID 는 자동 증가되는 스케쥴의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ActivationFlag 는 스케쥴의 활성화여부이다.
	ActivationFlag bool `json:"activation_flag,omitempty"`

	// Topic 은 스케쥴 실행시 Publishing 할 Topic 이다.
	Topic string `gorm:"not null" json:"topic,omitempty"`

	// Topic 은 스케쥴 실행시 Publishing 할 Message 이다.
	Message string `gorm:"not null" json:"message,omitempty"`

	// StartAt 은 스케쥴이 시작될 일시(UnixTime)이다.
	// 특정일시 스케쥴의 경우 스케쥴이 실행될 날짜를 나타낸다.
	StartAt int64 `gorm:"not null" json:"start_at,omitempty"`

	// EndAt 은 스케쥴이 종료될 일시(UnixTime)이다.
	EndAt int64 `gorm:"not null" json:"end_at,omitempty"`

	// Type 은 스케쥴의 종류이다.
	Type string `gorm:"not null" json:"type,omitempty"`

	// IntervalMinute 은 분 단위 스케쥴에서 스케쥴이 실행될 분 간격이다.
	IntervalMinute *uint `json:"interval_minute,omitempty"`

	// IntervalHour 는 시간 단위 스케쥴에서 스케쥴이 실행될 시간 간격이다.
	IntervalHour *uint `json:"interval_hour,omitempty"`

	// IntervalDay 는 일 단위 스케쥴에서 스케쥴이 실행될 일 간격이다.
	IntervalDay *uint `json:"interval_day,omitempty"`

	// IntervalWeek 는 주 단위 스케쥴에서 스케쥴이 실행될 주 간격이다.
	IntervalWeek *uint `json:"interval_week,omitempty"`

	// IntervalMonth 는 월 단위 스케쥴에서 스케쥴이 실행될 월 간격이다.
	IntervalMonth *uint `json:"interval_month,omitempty"`

	// WeekOfMonth 는 월 단위 스케쥴에서 스케쥴이 실행될 주차이다.
	WeekOfMonth *string `json:"week_of_month,omitempty"`

	// DayOfMonth 는 월 단위 스케쥴에서 스케쥴이 실행될 일자이다.
	DayOfMonth *string `json:"day_of_month,omitempty"`

	// DayOfWeek 는 주 단위 스케쥴에서 스케쥴이 실행될 요일이다.
	// 월 단위 스케쥴에서는 스케쥴이 실행될 요일을 나타낸다.
	DayOfWeek *string `json:"day_of_week,omitempty"`

	// Hour 는 스케쥴이 실행될 시간이다.
	// 시간단위, 분단위 스케쥴에서는 사용하지 않는다.
	Hour *uint `json:"hour,omitempty"`

	// Minute 는 스케쥴이 실행될 분이다.
	// 분단위 스케쥴에서는 사용하지 않는다.
	Minute *uint `json:"minute,omitempty"`

	// Timezone 은 스케쥴의 기준이 되는 타임존이다.
	Timezone string `gorm:"not null" json:"timezone,omitempty"`
}

// Backup 백업
type Backup struct {
	// ID 는 자동 증가되는 백업본의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// Path 는 백업본의 파일 위치이다.
	Path string `gorm:"not null" json:"path,omitempty"`

	// Remarks 는 백업본의 비고이다.
	Remarks *string `json:"remarks,omitempty"`

	// CreatedAt 는 백업본이 생성된 시간(UnixTime)이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// UpdatedAt 는 백업본 정보가 수정된 최종시간(UnixTime)이다.
	UpdatedAt int64 `gorm:"not null" json:"updated_at,omitempty"`
}

// GlobalConfig 전역 설정
type GlobalConfig struct {
	// Key 는 전역 설정 키이다.
	Key string `gorm:"primary_key" json:"key,omitempty"`

	// Value 는 전역 설정 값이다.
	Value string `gorm:"not null" json:"value,omitempty"`
}

// TenantConfig 테넌트 설정
type TenantConfig struct {
	// TenantID 는 테넌트의 ID 이다.
	TenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"tenant_id,omitempty"`

	// Key 는 테넌트 설정 키이다.
	Key string `gorm:"primary_key" json:"key,omitempty"`

	// Value 는 테넌트 설정 값이다.
	Value string `gorm:"not null" json:"value,omitempty"`
}

// ServiceConfig 서비스 설정
type ServiceConfig struct {
	// Name 은 서비스명이다.
	Name string `gorm:"primary_key" json:"name,omitempty"`

	// Key 는 서비스 설정 키이다.
	Key string `gorm:"primary_key" json:"key,omitempty"`

	// Value 는 서비스 설정 값이다.
	Value string `gorm:"not null" json:"value,omitempty"`
}

// TableName Tenant 테이블명 반환 함수
func (Tenant) TableName() string {
	return "cdm_tenant"
}

// BeforeCreate Tenant Record 생성 전 호출되는 함수
func (t *Tenant) BeforeCreate() error {
	t.CreatedAt = time.Now().Unix()
	t.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate Tenant Record 수정 전 호출되는 함수
func (t *Tenant) BeforeUpdate() error {
	t.UpdatedAt = time.Now().Unix()
	return nil
}

// TableName TenantSolution 테이블명 반환 함수
func (TenantSolution) TableName() string {
	return "cdm_tenant_solution"
}

// TableName TenantReceiveEvent 테이블명 반환 함수
func (TenantReceiveEvent) TableName() string {
	return "cdm_tenant_receive_event"
}

// TableName Role 테이블명 반환 함수
func (Role) TableName() string {
	return "cdm_role"
}

// TableName Group 테이블명 반환 함수
func (Group) TableName() string {
	return "cdm_group"
}

// BeforeCreate Group Record 생성 전 호출되는 함수
func (g *Group) BeforeCreate() error {
	g.CreatedAt = time.Now().Unix()
	g.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate Group Record 수정 전 호출되는 함수
func (g *Group) BeforeUpdate() error {
	g.UpdatedAt = time.Now().Unix()
	return nil
}

// TableName User 테이블명 반환 함수
func (User) TableName() string {
	return "cdm_user"
}

// BeforeCreate User Record 생성 전 호출되는 함수
func (u *User) BeforeCreate() error {
	u.CreatedAt = time.Now().Unix()
	u.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate User Record 수정 전 호출되는 함수
func (u *User) BeforeUpdate() error {
	u.UpdatedAt = time.Now().Unix()
	return nil
}

// TableName UserRole 테이블명 반환 함수
func (UserRole) TableName() string {
	return "cdm_user_role"
}

// TableName UserGroup 테이블명 반환 함수
func (UserGroup) TableName() string {
	return "cdm_user_group"
}

// TableName UserReceiveEvent 테이블명 반환 함수
func (UserReceiveEvent) TableName() string {
	return "cdm_user_receive_event"
}

// TableName Event 테이블명 반환 함수
func (Event) TableName() string {
	return "cdm_event"
}

// BeforeCreate Event Record 생성 전 호출되는 함수
func (e *Event) BeforeCreate() error {
	if e.CreatedAt == 0 {
		e.CreatedAt = time.Now().Unix()
	}
	return nil
}

// TableName EventCode 테이블명 반환 함수
func (EventCode) TableName() string {
	return "cdm_event_code"
}

// TableName EventCodeMessage 테이블명 반환 함수
func (EventCodeMessage) TableName() string {
	return "cdm_event_code_message"
}

// TableName EventError 테이블명 반환 함수
func (EventError) TableName() string {
	return "cdm_event_error"
}

// TableName Schedule 테이블명 반환 함수
func (Schedule) TableName() string {
	return "cdm_schedule"
}

// TableName Backup 테이블명 반환 함수
func (Backup) TableName() string {
	return "cdm_backup"
}

// BeforeCreate Backup Record 생성 전 호출되는 함수
func (b *Backup) BeforeCreate() error {
	b.CreatedAt = time.Now().Unix()
	b.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate Backup Record 수정 전 호출되는 함수
func (b *Backup) BeforeUpdate() error {
	b.UpdatedAt = time.Now().Unix()
	return nil
}

// TableName GlobalConfig 테이블명 반환 함수
func (GlobalConfig) TableName() string {
	return "cdm_global_config"
}

// TableName TenantConfig 테이블명 반환 함수
func (TenantConfig) TableName() string {
	return "cdm_tenant_config"
}

// TableName ServiceConfig 테이블명 반환 함수
func (ServiceConfig) TableName() string {
	return "cdm_service_config"
}
