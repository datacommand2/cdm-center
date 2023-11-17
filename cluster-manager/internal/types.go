package internal

// OpenstackCredential 클러스터 인증정보
type OpenstackCredential struct {
	Methods  []string              `json:"methods,omitempty"`
	Token    OpenstackTokenAuth    `json:"token,omitempty"`
	Password OpenstackPasswordAuth `json:"password,omitempty"`
}

// OpenstackTokenAuth 클러스터 토큰 인증
type OpenstackTokenAuth struct {
	ID string `json:"id,omitempty"`
}

// OpenstackPasswordAuth 클러스터 비밀번호 인증
type OpenstackPasswordAuth struct {
	User ClusterPasswordAuthUser `json:"user,omitempty"`
}

// ClusterPasswordAuthUser 클러스터 인증 사용자
type ClusterPasswordAuthUser struct {
	Name     string                        `json:"name,omitempty"`
	Domain   ClusterPasswordAuthUserDomain `json:"domain,omitempty"`
	Password string                        `json:"password,omitempty"`
}

// ClusterPasswordAuthUserDomain 클러스터 인증 사용자의 도메인
type ClusterPasswordAuthUserDomain struct {
	Name string `json:"name,omitempty"`
}
