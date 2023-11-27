package store

import "errors"

// DefaultStore 는 store 패키지의 기본 Store 이며, 지정하지 않으면 아무 기능도 수행하지 않는 noopStore 로 생성된다.
var DefaultStore Store

// ErrNotFoundKey 는 키를 찾을 수 없을 때, 발생하는 에러
var ErrNotFoundKey = errors.New("not found key")

// Store key, value data store 인터페이스
type Store interface {
	Options() Options
	Get(key string, opts ...GetOption) (string, error)
	GetAll(key string, opts ...GetOption) (string, error)
	Put(key, value string, opts ...PutOption) error
	Delete(key string, opts ...DeleteOption) error
	List(key string, opts ...ListOption) ([]string, error)
	Transaction(fc func(Txn) error, opts ...TxnOption) error
	Close() error
	Connect() error
}

// Txn Key-Value Store 의 트랜잭션
type Txn interface {
	Put(key, value string)
	Delete(key string, opts ...DeleteOption)
}

// Get 은 DefaultStore 에서 key 에 해당하는 value 를 반환하는 함수이다.
// 해당 key 가 존재하지 않을 경우 빈값과 에러가 리턴된다.
func Get(key string, opts ...GetOption) (string, error) {
	return DefaultStore.Get(key, opts...)
}

// GetAll 은 DefaultStore 에서 key 에 해당하는 value 를 반환하는 함수이다.
// 해당 key 가 존재하지 않을 경우 빈값과 에러가 리턴된다.
func GetAll(key string, opts ...GetOption) (string, error) {
	return DefaultStore.GetAll(key, opts...)
}

// Put 은 DefaultStore 에 key, value 를 저장하는 함수이다.
// 이미 key 가 존재할 경우 덮어쓴다.
func Put(key, value string, opts ...PutOption) error {
	return DefaultStore.Put(key, value, opts...)
}

// Delete 는 DefaultStore 에서 key value 를 삭제하는 함수이다.
// key 가 존재하지 않더라도 에러를 리턴하지 않는다.
// DeletePrefix() 옵션 설정시 해당 key 가 prefix 인 key value 목록 전체를 삭제한다.
func Delete(key string, opts ...DeleteOption) error {
	return DefaultStore.Delete(key, opts...)
}

// List 는 DefaultStore 에 해당 key 가 prefix 인 모든 key 목록을 조회하는 함수이다.
func List(key string, opts ...ListOption) ([]string, error) {
	return DefaultStore.List(key, opts...)
}

// Close 는 DefaultStore 와의 연결을 종료하는 함수이다.
func Close() error {
	return DefaultStore.Close()
}

// Connect 는 DefaultStore 로 연결 하는 함수이다.
func Connect() error {
	return DefaultStore.Connect()
}

// Init 는 DefaultStore 을 etcd store 로 설정 하는 함수이다
func Init(name string, opts ...Option) {
	DefaultStore = NewStore(name, opts...)
}

// Transaction 은 DefaultStore 의 트랜잭션 함수이다.
func Transaction(fc func(Txn) error, opts ...TxnOption) error {
	return DefaultStore.Transaction(fc, opts...)
}
