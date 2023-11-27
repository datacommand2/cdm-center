package database

import (
	"context"
	"fmt"
	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

// ConnectionWrapper 는 지속적으로 데이터베이스 접속을 유지하기 위한 gorm.DB 의 래퍼 구조체이다.
type ConnectionWrapper struct {
	*gorm.DB
	tx     *gorm.DB
	txLock sync.Mutex

	opts  Options
	close chan struct{}
}

func (c *ConnectionWrapper) dataSource(n *registry.Node) (string, string, error) {
	h, p, _ := net.SplitHostPort(n.Address)

	// dialect
	d, ok := n.Metadata["dialect"]
	if !ok {
		return "", "", errors.New("not found 'dialect' in database service node")
	}

	// ssl connection
	var opts []string
	if c.opts.SSLEnable {
		switch d {
		case "postgres":
			opts = append(opts, "sslmode=verify-ca")

		default:
			logger.Warnf("SSL Connection is unsupported. database dialect(%s)", d)
		}

	} else {
		switch d {
		case "postgres":
			opts = append(opts, "sslmode=disable")
		}
	}

	if c.opts.SSLEnable && c.opts.SSLCACert != "" {
		switch d {
		case "postgres":
			opts = append(opts, fmt.Sprintf("sslrootcert=%s", c.opts.SSLCACert))

		default:
			logger.Warnf("SSL Connection is unsupported. database dialect(%s)", d)
		}
	}

	// data source
	var s string
	switch d {
	case "mysql":
		s = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			c.opts.Username, c.opts.Password, h, p, c.opts.DBName, strings.Join(opts, "&"))

	case "postgres":
		s = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s %s",
			h, p, c.opts.Username, c.opts.DBName, c.opts.Password, strings.Join(opts, " "))

	default:
		return "", "", errors.New("unsupported database dialect")
	}

	return d, s, nil
}

// connect 는 데이터베이스 서버에 대한 서비스를 찾고, 접속을 시도하는 함수이다.
func (c *ConnectionWrapper) connect() error {
	// discover service
	var r registry.Registry
	if c.opts.Registry == nil {
		r = registry.DefaultRegistry
	} else {
		r = c.opts.Registry
	}

	logger.Debugf("Discovering database(%s) service.", c.opts.ServiceName)
	s, err := r.GetService(c.opts.ServiceName)
	switch {
	case err != nil && err != registry.ErrNotFound:
		logger.Errorf("Could not discover database(%s) service. cause: %v", c.opts.ServiceName, err)
		return errors.New("could not discover database service")

	case err == registry.ErrNotFound || s == nil || len(s) == 0:
		logger.Errorf("Not found database(%s) service.", c.opts.ServiceName)
		return errors.New("not found database service")
	}

	// loop service nodes
	logger.Debugf("Connecting to database(%s) service.", c.opts.ServiceName)
	next := selector.RoundRobin(s)
	for i := 0; i < len(s[0].Nodes); i++ {
		n, _ := next()
		d, s, err := c.dataSource(n)
		if err != nil {
			logger.Warnf("Unable to make DataSource. cause: %v", err)
			continue
		}

		// connect
		logger.Tracef("Connecting to database(dialect: %s, address: %s)", d, n.Address)
		db, err := gorm.Open(d, s)
		if err != nil {
			logger.Tracef("Database(dialect: %s, address: %s) connection failed. cause: %v", d, n.Address, err)
			continue
		}

		logger.Infof("Database(dialect: %s, address: %s) connected", d, n.Address)

		old := c.DB
		c.DB = db

		if old != nil {
			_ = old.Close()
		}

		return nil
	}

	// connection failed all service node
	return errors.New("could not connect to database")
}

// monitorHealthStatus 는 데이터베이스 서버와의 접속여부를 주기적으로 모니터링하는 함수이다.
func (c *ConnectionWrapper) monitorHealthStatus(ch chan struct{}) {
	for {
		time.Sleep(c.opts.HeartbeatInterval)

		// health check
		if c.DB.DB().Ping() != nil {
			close(ch)
			break
		}
	}
}

// reconnect 는 데이터베이스 서버와의 접속이 끊기면, 재접속을 시도하는 함수이다.
func (c *ConnectionWrapper) reconnect() {
	var conn = false

	for {
		if conn {
			// connect
			logger.Warnf("Try to reconnect database(%s) service.", c.opts.ServiceName)
			if err := c.connect(); err != nil {
				// sleep when connection failed
				time.Sleep(c.opts.ReconnectInterval)
				continue
			}
		}

		conn = true
		ch := make(chan struct{})

		// start health check thread
		logger.Debugf("Start database(%s) connection monitoring.", c.opts.ServiceName)
		go c.monitorHealthStatus(ch)

		select {
		case <-ch:
			// health check failed
			logger.Warnf("Database(%s) connection might be closed.", c.opts.ServiceName)
			continue

		case <-c.close:
			// closed by caller
			logger.Debugf("Stop databased(%s) connection monitoring.", c.opts.ServiceName)
			return
		}
	}
}

// Close 는 데이터베이스 접속을 종료하는 함수이다.
func (c *ConnectionWrapper) Close() error {
	select {
	case <-c.close:
		logger.Debugf("Database(%s) connection is already closed.", c.opts.ServiceName)
		return nil

	default:
		logger.Infof("Closing database(%s) connection.", c.opts.ServiceName)
		close(c.close)
	}

	return c.DB.Close()
}

// Test start a transaction as a test block,
// return error will rollback, otherwise to commit.
func (c *ConnectionWrapper) Test(fc func(db *gorm.DB)) {
	c.txLock.Lock()
	c.tx = c.Begin()
	c.txLock.Unlock()

	fc(c.tx)

	c.txLock.Lock()
	c.tx.Rollback()
	c.tx = nil
	c.txLock.Unlock()
}

// TODO: 아래 코드는 cockroach 용이며, 다른 DBMS 를 지원해야 할 경우 추가 개발이 필요하다.
func (c *ConnectionWrapper) nestedTransaction(fc func(*gorm.DB) error) (err error) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	savepoint := string(b)

	// savepoint
	if err := c.tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepoint)).Error; err != nil {
		return err
	}

	panicked := true
	defer func() {
		// rollback to savepoint
		if panicked || err != nil {
			c.tx.Exec(fmt.Sprintf("ROLLBACK TO %s", savepoint))
		}
	}()

	// run transactional block
	err = fc(c.tx)

	// release savepoint
	if err == nil {
		c.tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepoint))
	}

	panicked = false
	return
}

// Transaction start a transaction as a block,
// return error will rollback, otherwise to commit.
func (c *ConnectionWrapper) Transaction(fc func(*gorm.DB) error) error {
	if c.opts.TestMode {
		c.txLock.Lock()
		defer c.txLock.Unlock()

		if c.tx == nil {
			return errors.New("test transaction is not started")
		}

		return c.nestedTransaction(fc)
	}
	// ExecuteInTx 함수에서 40001(serialization_failure) error 일 경우 계속 트랜 잭션을 수행 하며,
	// 트랜잭션이 concurrency test 를 수행 할 경우에 에러가 발생 함. 일반적인 경우 발생 할 가능성이 적음
	// 일반 적인 경우에도 트랜잭션 에러가 발생 할 경우, Context timeout 추가 할 필요가 있음
	ctx := context.Background()
	tx := c.DB.BeginTx(ctx, nil)

	var err error
	if err = tx.Error; err != nil {
		return errors.UnusableDatabase(err)
	}

	var txError error
	txError = crdb.ExecuteInTx(ctx, gormTxAdapter{tx}, func() error {
		err = fc(tx)
		if errors.Equal(err, errors.ErrUnusableDatabase) {
			return errors.UnwrapUnusableDatabase(err)
		}

		return err
	})

	if err == nil && txError != nil {
		return errors.UnusableDatabase(txError)
	}

	return err
}

// Execute NoTransaction start a Execute as a block,
func (c *ConnectionWrapper) Execute(fc func(*gorm.DB) error) error {
	if c.opts.TestMode {
		c.txLock.Lock()
		defer c.txLock.Unlock()

		if c.tx == nil {
			return errors.New("test transaction is not started")
		}

		return c.nestedTransaction(fc)
	}
	// Transaction 없이 진행
	var err error

	err = fc(c.DB)
	if errors.Equal(err, errors.ErrUnusableDatabase) {
		return errors.UnwrapUnusableDatabase(err)
	}

	return err
}

// GormTransaction start a transaction as a block,
// return error will rollback, otherwise to commit.
func (c *ConnectionWrapper) GormTransaction(fc func(*gorm.DB) error) error {
	if c.opts.TestMode {
		c.txLock.Lock()
		defer c.txLock.Unlock()

		if c.tx == nil {
			return errors.New("test transaction is not started")
		}

		return c.nestedTransaction(fc)
	}
	// Gorm Transaction 을 수행
	ctx := context.Background()
	tx := c.DB.BeginTx(ctx, nil)

	var err error
	err = fc(tx)

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if err = tx.Error; err != nil {
		return errors.UnusableDatabase(err)
	}

	return err
}

// gormTxAdapter adapts a *gorm.DB to a crdb.Tx.
type gormTxAdapter struct {
	db *gorm.DB
}

var _ crdb.Tx = gormTxAdapter{}

// Exec is part of the crdb.Tx interface.
func (tx gormTxAdapter) Exec(_ context.Context, q string, args ...interface{}) error {
	return tx.db.Exec(q, args...).Error
}

// Commit is part of the crdb.Tx interface.
func (tx gormTxAdapter) Commit(_ context.Context) error {
	return tx.db.Commit().Error
}

// Rollback is part of the crdb.Tx interface.
func (tx gormTxAdapter) Rollback(_ context.Context) error {
	return tx.db.Rollback().Error
}
