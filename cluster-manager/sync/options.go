package sync

// Options sync 시 사용하는 옵션 구조체
type Options struct {
	Force bool

	TenantUUID string
}

// Option 옵션 함수
type Option func(*Options)

// Force sync 를 강제로 할지 여부를 결정하는 옵션 함수
func Force(f bool) Option {
	return func(o *Options) {
		o.Force = f
	}
}

// TenantUUID tenant uuid 정보
func TenantUUID(u string) Option {
	return func(o *Options) {
		o.TenantUUID = u
	}
}
