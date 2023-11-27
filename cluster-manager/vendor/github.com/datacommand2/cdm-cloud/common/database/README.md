# Database

CDM-Cloud 의 공통 데이터베이스 서버에 접속하는 기능을 구현한다.


## Usage
Database 라이브러리를 사용하기 위해 다음 패키지를 import 한다.
```go
import "github.com/datacommand2/cdm-cloud/common/database"
```

### 초기화
Database 에 접속하기 위한 기본 설정을 초기화한 뒤 사용할 수 있다. 초기화 함수는 다음과 같다.

```go
// Init 은 데이터베이스 서버를 찾고, 접속하기 위한 기본 설정 값을 초기화하는 함수이다.
// serviceName 은 Service Registry 에 등록된 데이터베이스 서버에 대한 서비스 이름이며, 서비스 이름을 통해 registry 에서 접속주소를 찾는다.
// dbName, username, password 는 접속할 데이터베이스 이름, 계정, 비밀번호이다.
func Init(serviceName, dbName, username, password string, opts ...Option)
```

추가 옵션을 지정하여 기본 설정 값을 초기화할 수 있으며, 설정할 수 있는 옵션은 다음과 같다.

```go
// Registry 는 서비스 레지스트리를 설정하는 함수이다.
// 설정하지 않을 경우 registry.DefaultRegistry 를 사용한다.
func Registry(r registry.Registry) Option

// SSLEnable 은 SSL 접속 활성화 여부를 설정하는 함수이다.
func SSLEnable(enable bool) Option

// SSLCACert 는 서버에 적용된 인증서를 발행한 CA 의 인증서를 설정하는 함수이다.
func SSLCACert(ca string) Option

// HeartbeatInterval 은 접속유지 여부 확인 주기를 설정하는 함수이며, 설정시 기본 설정의 확인 주기를 대체한다.
func HeartbeatInterval(i time.Duration) Option

// ReconnectInterval 은 재접속 시도 주기를 설정하는 함수이며, 설정시 기본 설정의 시도 주기를 대체한다.
func ReconnectInterval(i time.Duration) Option

// TestMode 는 테스트 모드로 설정하는 함수이며, 테스트 모드에서의 변경분은 모두 롤백된다.
func TestMode() Option
```

### 데이터베이스 접속 및 접속 종료
기본 데이터베이스에 접속하거나 접속을 종료하기 위해 제공하는 함수는 다음과 같다.
```go
// Connect 은 기본 설정으로 데이터베이스 서버에 접속하는 함수이다.
// 접속 성공시 데이터베이스 서버와의 접속여부를 주기적으로 모니터링하고 끊겼을 경우, 유효한 데이터베이스 서버에 재접속하는 기능이 수행된다.
// 유효한 데이터베이스 서버가 없더라도 데이터베이스 서버가 재기동될 때까지 재접속을 시도하며,
// 재접속에 성공하기 전까지는 Transaction 함수 호출 시 'connection refused' 가 발생할 수 있다.
func Connect() error

// Close 는 기본 설정으로 접속한 데이터베이스의 접속을 종료하는 함수이다.
func Close() error
```

### 데이터베이스 사용
기본 데이터베이스는 무조건 트랜잭션으로 처리해야 하며, 다음과 같이 사용할 수 있다.
```go
// Transaction 은 기본 설정으로 접속한 데이터베이스의 트랜잭션 함수이다.
database.Transaction(func(tx *gorm.DB) error {
	tx.Save(...)
	
	if wantRollback {
        return errors.New("rollback !!")
    } else if wantCommit {
        return nil // commit !!
    }
})
```

### 테스트 모드
다음은 테스트 모드로 설정한 경우 Transaction(...) 함수를 호출하기위한 example 이다.
```go
t.Log(v) // 'aaa'

database.Test(func(db *gorm.DB) {
    v = "bbb"
    db.Save(&v)

    db.Find(&v)
    t.Log(v) // 'bbb'

    database.Transaction(func(tx *gorm.DB) error {
    	v = "ccc"
        tx.Save(&v)
        return errors.New("rollback !!")
    })

    db.Find(&v)
    t.Log(v) // 'bbb'

    database.Transaction(func(tx *gorm.DB) error {
        v = "ddd"
        tx.Save(&v)
        return nil // commit !!
    })

    db.Find(&v)
    t.Log(v) // 'ddd'
})

db.Find(&v)
t.Log(v) // 'aaa'
```

### 다른 데이터베이스
기본 데이터베이스 외에 다른 데이터베이스에 접속할 수 있으며, 접속 및 접속종료, 트랜잭션 함수는 다음과 같다.
```go
// Open 은 데이터베이스 서버에 접속하는 함수이다.
// serviceName 은 Service Registry 에 등록된 데이터베이스 서버에 대한 서비스 이름이며, 서비스 이름을 통해 registry 에서 접속주소를 찾는다.
// dbName, username, password 는 접속할 데이터베이스 이름, 계정, 비밀번호이다.
// 접속에 성공하면 *ConnectionWrapper 를 반환하며, 사용한 뒤 Close() 를 호출해서 닫아줘야 한다.
//
// 접속 성공시 데이터베이스 서버와의 접속여부를 주기적으로 모니터링하고 끊겼을 경우, 유효한 데이터베이스 서버에 재접속하는 기능이 수행된다.
// 유효한 데이터베이스 서버가 없더라도 데이터베이스 서버가 재기동될 때까지 재접속을 시도하며,
// 재접속에 성공하기 전까지는 Transaction 함수 호출 시 'connection refused' 가 발생할 수 있다.
func Open(serviceName, dbName, username, password string, opts ...Option) (*ConnectionWrapper, error)

// Close 는 데이터베이스 접속을 종료하는 함수이다.
func (c *ConnectionWrapper) Close() error

// Transaction start a transaction as a block,
// return error will rollback, otherwise to commit.
func (c *ConnectionWrapper) Transaction(fc func(*gorm.DB) error) (err error)
```
