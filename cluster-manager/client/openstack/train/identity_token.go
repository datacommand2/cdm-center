package train

import "time"

// Token 토큰 구조체
type Token struct {
	Key     string  `json:"-"`
	Payload Payload `json:"token"`
}

// Role 역할 구조체
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Domain 구조체
type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// System 시스템 구조체
type System struct {
	All bool `json:"all"`
}

// Project 프로젝트 구조체
type Project struct {
	Domain Domain `json:"domain"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

// Endpoint 엔드 포인트 구조체
type Endpoint struct {
	RegionID  string `json:"region_id"`
	URL       string `json:"url"`
	Region    string `json:"region"`
	Interface string `json:"interface"`
	ID        string `json:"id"`
}

// Catalog 카탈로그 구조체
type Catalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	Type      string     `json:"type"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
}

// User 사용자 구조체
type User struct {
	PasswordExpiresAt string `json:"password_expires_at"`
	Domain            Domain `json:"domain"`
	ID                string `json:"id"`
	Name              string `json:"name"`
}

// Payload 토큰의 정보를 저장하는 구조체
type Payload struct {
	IsDomain                        bool       `json:"is_domain"`
	Methods                         []string   `json:"methods"`
	Domain                          Domain     `json:"domain"`
	System                          System     `json:"system"`
	Roles                           []Role     `json:"roles"`
	ExpiresAt                       time.Time  `json:"expires_at"`
	Project                         Project    `json:"project"`
	Catalog                         []Catalog  `json:"catalog"`
	ID                              string     `json:"id,omitempty"`
	Name                            string     `json:"name,omitempty"`
	ApplicationCredentialRestricted bool       `json:"application_credential_restricted"`
	User                            User       `json:"user"`
	AuditIds                        []string   `json:"audit_ids"`
	IssuedAt                        time.Time  `json:"issued_at"`
	RegionID                        string     `json:"region_id"`
	URL                             string     `json:"url"`
	Region                          string     `json:"region"`
	Interface                       string     `json:"interface"`
	Endpoints                       []Endpoint `json:"endpoints"`
	Type                            string     `json:"type"`
}
