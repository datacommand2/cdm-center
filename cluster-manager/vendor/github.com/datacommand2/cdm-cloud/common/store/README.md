store
----------
  * key, value store를 사용하기 위한 common의 공용 라이브러리
  * go-micro store interface를 참조 하여 수정함
  * go-micro store option값을 참조하여 수정함
    * Options - Node
    * PutOptions - PutTimeout 추가, PutTTL 사용
    * GetOptions - GetTimeout 추가
    * DeleteOptions - DeleteTimeout 추가, DeletePrefix 추가
    * ListOptions - ListTimeout 추가

store/etcd
----------
  * ETCD를 사용하기 위한 common의 공용 라이브러리
  * store.go 인터페이스를 구현

### 패키지
```go
import "github.com/datacommand2/cdm-cloud/common/store/store"
import "github.com/datacommand2/cdm-cloud/common/store/store/etcd"
```

### store/store.go 
```go
type Store interface {
	// Options NewStore()로 설정된 Store의 옵션 조회.
	Options() Options
	// Get Key에 대한 Value를 조회함.
	Get(key string, opts ...GetOption) (*string, error)
	// Put Key, Value를 저장함.
	Put(key, value string, opts ...PutOption) error
	// Delete Key에 해당하는 Key, Value를 삭제 함.
	// 옵션 값의 따라 prefix에 해당하는 모든 Key, Value를 삭제 가능
	Delete(key string, opts ...DeleteOption) error
	// List Key를 prefix로 하여 일치하는 모든 키를 조회 함.
	List(key string, opts ...ListOption) ([]string, error)
	// Close NewStore()로 생성된 Store 종료
	Close() error
}

```
### store/option.go 
```go
// Registry 는 서비스 레지스트리를 설정하는 함수이다.
// 설정하지 않을 경우 registry.DefaultRegistry 를 사용한다.
func Registry(r registry.Registry) Option

// ServiceName 은 서비스 이름을 설정하는 함수이다.
// Service Registry 에 등록된 Key-Value Store 서버에 대한 서비스 이름이어야 한다.
func ServiceName(name string) Option

// GetTimeout Get 요청 타임아웃 설정
func GetTimeout(d time.Duration) GetOption

// PutTTL 데이터 만료 시간 설정
func PutTTL(d time.Duration) PutOption
// PutTimeout Put 요청 타임아웃 설정
func PutTimeout(d time.Duration) PutOption

//Deleteoptions- timeout, prefix 설정 
func DeleteTimeout(t time.Duration) DeleteOption
func DeletePrefix() DeleteOption

//ListOptions- timeout 설정 
func ListTimeout(d time.Duration) ListOption
```
### store/etcd/etcd.go 
```go
//defaltTimeout - 옵션에서 Timeout 값이 설정되지 않을 경우 사용하는 디폴트값
var defaltTimeout = 5* time.Second

// etcdStore -ETCD 를 사용하기 위한 구조체
type etcdStore struct {
	//options -Store의 옵션
	options store.Options
	//client -ETCD client 구조체
	client *clientv3.Client
	//config -ETCD client 설정값을 위한 config 구조체
	config clientv3.Config
}

// NewStore etcdStore 생성 및 반환
func NewStore(opts ...store.Option) (store.Store, error)

// Close etcdStore.client의 연결 종료
func (e *etcdStore) Close() error

// Options 현재 설정된 etcdStore의 옵션을 반환
func (e *etcdStore) Options() store.Options

// Get key로 저장된 value를 조회
// 옵션으로 timeout 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) Get(key string, opts ...store.GetOption) (*string, error)

// Put key, value를 저장
// 옵션으로 timeout, TTL 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) Put(key, value string, opts ...store.PutOption) error

// Delete key로 해당 key, value를 삭제
// 옵션으로 timeout, Prefix 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
// prefix 옵션값으로 지정되는 경우 prefix에 일치하는 모든 key, value를 삭제함
func (e *etcdStore) Delete(key string, opts ...store.DeleteOption) error

// List key를 preifx로 하여 일치하는 모든 key를 조회
// 옵션으로 timeout이 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) List(key string, opts ...store.ListOption) ([]string, error)

``` 
### Usage
---------
```go
//생성
store := etcd.NewStore(store.Nodes(endpoint ...string),store.option() ...) //etcd config.go 참고 
key, value:= "..", "..."

//Put
err := store.Put(key, value, store.PutTimeout(5 * time.Second), store.PutTTL(5 * time.Second))

//Get
Value, err := store.Read(key, store.GetTimeout(5 * time.Second))

//Delete
err := store.Delete(key, store.DeletePrefix(), store.DeleteTimeout(5 * time.Second))

//List
keylist, err := store.List(key, store.ListTimeout(5 * time.Second))
```