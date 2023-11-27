package model

import (
	"github.com/jinzhu/gorm"
)

// Groups 는 사용자의 그룹 목록을 조회하는 함수이다.
func (u *User) Groups(db *gorm.DB) ([]Group, error) {
	var relations []UserGroup
	if err := db.Where("user_id=?", u.ID).Find(&relations).Error; err != nil {
		return nil, err
	}

	var arr []uint64
	for _, rel := range relations {
		arr = append(arr, rel.GroupID)
	}

	var groups []Group
	if err := db.Where("id IN (?)", arr).Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// Roles 는 사용자의 솔루션 역할 목록을 조회하는 함수이다.
func (u *User) Roles(db *gorm.DB) ([]Role, error) {
	var relations []UserRole
	if err := db.Where("user_id=?", u.ID).Find(&relations).Error; err != nil {
		return nil, err
	}

	var arr []uint64
	for _, rel := range relations {
		arr = append(arr, rel.RoleID)
	}

	var roles []Role
	if err := db.Where("id IN (?)", arr).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

// Tenant 는 이벤트의 테넌트를 조회하는 함수이다.
func (e *Event) Tenant(db *gorm.DB) (*Tenant, error) {
	var tenant Tenant

	if err := db.Where("id=?", e.TenantID).Find(&tenant).Error; err != nil {
		return nil, err
	}

	return &tenant, nil
}

// EventCode 는 이벤트의 이벤트 코드를 조회하는 함수이다.
func (e *Event) EventCode(db *gorm.DB) (*EventCode, error) {
	var eventCode EventCode

	if err := db.Where("code=?", e.Code).Find(&eventCode).Error; err != nil {
		return nil, err
	}

	return &eventCode, nil
}

// EventError 는 이벤트의 에러 코드를 조회하는 함수이다.
func (e *Event) EventError(db *gorm.DB) (*EventError, error) {
	var eventError EventError

	if err := db.Where("code=?", e.ErrorCode).Find(&eventError).Error; err != nil {
		return nil, err
	}

	return &eventError, nil
}
