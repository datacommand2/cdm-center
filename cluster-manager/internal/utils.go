package internal

import identity "github.com/datacommand2/cdm-cloud/services/identity/proto"

// IsAdminUser admin 역할의 사용자인지 확인
func IsAdminUser(user *identity.User) bool {
	for _, r := range user.GetRoles() {
		if r.Role == "admin" {
			return true
		}
	}
	return false
}

// IsGroupUser 그룹에 속한 사용자인지 확인
func IsGroupUser(user *identity.User, groupID uint64) bool {
	for _, g := range user.GetGroups() {
		if groupID == g.Id {
			return true
		}
	}
	return false
}
