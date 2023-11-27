package database

import (
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"time"
	// mysql dialect
	_ "github.com/jinzhu/gorm/dialects/mysql"
	// postgres dialect
	_ "github.com/lib/pq"
)

var defaultDatabase *ConnectionWrapper

// Init 은 데이터베이스 서버를 찾고, 접속하기 위한 기본 설정 값을 초기화하는 함수이다.
// serviceName 은 Service Registry 에 등록된 데이터베이스 서버에 대한 서비스 이름이며, 서비스 이름을 통해 registry 에서 접속주소를 찾는다.
// dbName, username, password 는 접속할 데이터베이스 이름, 계정, 비밀번호이다.
func Init(serviceName, dbName, username, password string, opts ...Option) {
	var options = Options{
		ServiceName:       serviceName,
		DBName:            dbName,
		Username:          username,
		Password:          password,
		HeartbeatInterval: 10 * time.Second,
		ReconnectInterval: 10 * time.Second,
	}

	for _, o := range opts {
		o(&options)
	}

	defaultDatabase = &ConnectionWrapper{
		opts:  options,
		close: make(chan struct{}),
	}
}

// OpenDefault 는 cli flag 및 추가 적인 옵션으로 데이터 베이스 서버에 접속하는 함수이다.
// 접속에 성공하면 *ConnectionWrapper 를 반환하며, 사용한 뒤 Close() 를 호출해서 닫아줘야 한다.
func OpenDefault(opts ...Option) (*ConnectionWrapper, error) {
	var options Options
	if err := copier.Copy(&options, &defaultDatabase.opts); err != nil {
		return nil, err
	}

	for _, o := range opts {
		o(&options)
	}

	db := &ConnectionWrapper{
		opts:  options,
		close: make(chan struct{}),
	}

	if err := db.connect(); err != nil {
		return nil, err
	}

	go db.reconnect()

	return db, nil
}

// Open 은 데이터베이스 서버에 접속하는 함수이다.
// serviceName 은 Service Registry 에 등록된 데이터베이스 서버에 대한 서비스 이름이며, 서비스 이름을 통해 registry 에서 접속주소를 찾는다.
// dbName, username, password 는 접속할 데이터베이스 이름, 계정, 비밀번호이다.
// 접속에 성공하면 *ConnectionWrapper 를 반환하며, 사용한 뒤 Close() 를 호출해서 닫아줘야 한다.
//
// 접속 성공시 데이터베이스 서버와의 접속여부를 주기적으로 모니터링하고 끊겼을 경우, 유효한 데이터베이스 서버에 재접속하는 기능이 수행된다.
// 유효한 데이터베이스 서버가 없더라도 데이터베이스 서버가 재기동될 때까지 재접속을 시도하며,
// 재접속에 성공하기 전까지는 Transaction 함수 호출 시 'connection refused' 가 발생할 수 있다.
func Open(serviceName, dbName, username, password string, opts ...Option) (*ConnectionWrapper, error) {
	var options = Options{
		ServiceName:       serviceName,
		DBName:            dbName,
		Username:          username,
		Password:          password,
		HeartbeatInterval: 10 * time.Second,
		ReconnectInterval: 10 * time.Second,
	}

	for _, o := range opts {
		o(&options)
	}

	w := &ConnectionWrapper{
		opts:  options,
		close: make(chan struct{}),
	}

	if err := w.connect(); err != nil {
		logger.Errorf("Could not connect to database.")
		return nil, err
	}

	go w.reconnect()

	return w, nil
}

// Connect 은 기본 설정으로 데이터베이스 서버에 접속하는 함수이다.
// 접속 성공시 데이터베이스 서버와의 접속여부를 주기적으로 모니터링하고 끊겼을 경우, 유효한 데이터베이스 서버에 재접속하는 기능이 수행된다.
// 유효한 데이터베이스 서버가 없더라도 데이터베이스 서버가 재기동될 때까지 재접속을 시도하며,
// 재접속에 성공하기 전까지는 Transaction 함수 호출 시 'connection refused' 가 발생할 수 있다.
func Connect() error {
	if err := defaultDatabase.connect(); err != nil {
		return err
	}

	go defaultDatabase.reconnect()

	return nil
}

// Close 는 기본 설정으로 접속한 데이터베이스의 접속을 종료하는 함수이다.
func Close() error {
	return defaultDatabase.Close()
}

// Test 는 테스트 모드에서 트랜잭션을 처리하는 함수이며, 함수내 트랜잭션 중첩을 방지한다.
func Test(fc func(*gorm.DB)) {
	defaultDatabase.Test(fc)
}

// Transaction 은 기본 설정으로 접속한 데이터베이스의 트랜잭션 함수이다.
func Transaction(fc func(*gorm.DB) error) error {
	return defaultDatabase.Transaction(fc)
}

// Execute 은 Transaction 을 사용하지 않는다.
func Execute(fc func(*gorm.DB) error) error {
	return defaultDatabase.Execute(fc)
}

// GormTransaction 은 Gorm 에서 제공해주는 Transaction 기능을 사용하는 함수이다.
func GormTransaction(fc func(*gorm.DB) error) error {
	return defaultDatabase.GormTransaction(fc)
}
