package common

import (
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/constant"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/logger/watcher"
	"github.com/datacommand2/cdm-cloud/common/store"
	"github.com/datacommand2/cdm-cloud/common/sync"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/config/cmd"
	mLogger "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
	"io"
	"os"
	"time"

	_ "github.com/go-micro/plugins/v2/registry/kubernetes"
)

type command struct {
	cmd.Cmd

	opts Options
}

var defaultPersistentQueues = []string{
	constant.QueueReportEvent,
}

var defaultFlags = []cli.Flag{
	// micro registry
	&cli.StringFlag{
		Name:     "micro_registry",
		Usage:    "Name of micro registry",
		Required: false,
	},

	// message broker (RabbitMQ)
	&cli.StringFlag{
		Name:     "cdm_broker_service_name",
		Usage:    "Name of cdm message broker service",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_broker_username",
		Usage:    "Username for authentication message broker",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_broker_password",
		Usage:    "Password for authentication message broker",
		Required: true,
	},

	// key-value store (ETCD)
	&cli.StringFlag{
		Name:     "cdm_store_service_name",
		Usage:    "Name of cdm key-value store service",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_store_username",
		Usage:    "Username for authentication key-value store",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "cdm_store_password",
		Usage:    "Password for authentication key-value store",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "cdm_store_heartbeat_interval",
		Usage:    "Interval of store connection check",
		Required: false,
	},

	// sync (ETCD)
	&cli.StringFlag{
		Name:     "cdm_sync_backend_service_name",
		Usage:    "Name of cdm sync backend service",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_sync_backend_username",
		Usage:    "Username for authentication sync backend",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "cdm_sync_backend_password",
		Usage:    "Password for authentication sync backend",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "cdm_sync_backend_heartbeat_interval",
		Usage:    "Interval of sync backend connection check",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "cdm_sync_ttl",
		Usage:    "sync backend expired time",
		Required: false,
	},

	// database
	&cli.StringFlag{
		Name:     "cdm_database_service_name",
		Usage:    "Name of cdm database service",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_database_name",
		Usage:    "Name of cdm database instance",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_database_username",
		Usage:    "Username for authentication database",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "cdm_database_password",
		Usage:    "Password for authentication database",
		Required: true,
	},
	&cli.BoolFlag{
		Name:     "cdm_database_ssl_enable",
		Usage:    "Enable SSL for database connection",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "cdm_database_ssl_ca",
		Usage:    "Certification file for database SSL connection",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "cdm_database_heartbeat_interval",
		Usage:    "Interval of database connection check",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "cdm_database_reconnect_interval",
		Usage:    "Interval of database re-connection",
		Required: false,
	},

	// micro framework
	&cli.StringFlag{
		Name:     "micro_logger_level",
		Value:    "warn",
		Usage:    "Default level for the micro framework logger",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "micro_client_request_timeout",
		Value:    30,
		Usage:    "Default request timeout for micro client",
		Required: false,
	},
}

func (c *command) initDefaults(ctx *cli.Context) error {
	// default micro client request timeout
	if timeout := ctx.Int("micro_client_request_timeout"); timeout > 0 {
		client.DefaultRequestTimeout = time.Second * time.Duration(timeout)
	}

	if registryName := ctx.String("micro_registry"); registryName != "" {
		fc, ok := cmd.DefaultRegistries[registryName]
		if !ok {
			return errors.New("invalid registry name")
		}
		registry.DefaultRegistry = fc()
	}

	logger.Infof("Registry(%s) running", registry.String())

	return nil
}

func (c *command) initBroker(ctx *cli.Context) error {
	// initialize default cdm message broker
	var opts []broker.Option
	if len(c.opts.PersistentQueues) > 0 {
		opts = append(opts, broker.PersistentQueue(c.opts.PersistentQueues...))
	}

	if err := broker.Init(
		ctx.String("cdm_broker_service_name"),
		ctx.String("cdm_broker_username"),
		ctx.String("cdm_broker_password"),
		opts...,
	); err != nil {
		return err
	}

	return broker.Connect()
}

func (c *command) destroyBroker() error {
	if broker.DefaultBroker == nil {
		return nil
	}

	return broker.Disconnect()
}

func (c *command) initStore(ctx *cli.Context) error {
	var opts []store.Option

	if ctx.IsSet("cdm_store_username") && ctx.IsSet("cdm_store_password") {
		opts = append(opts, store.Auth(ctx.String("cdm_store_username"), ctx.String("cdm_store_password")))
	}

	if ctx.IsSet("cdm_store_heartbeat_interval") {
		opts = append(opts, store.HeartbeatInterval(time.Duration(ctx.Int("cdm_store_heartbeat_interval"))*time.Second))
	}

	// initialize default cdm key-value store
	store.Init(ctx.String("cdm_store_service_name"), opts...)

	return store.Connect()
}

func (c *command) destroyStore() error {
	if store.DefaultStore == nil {
		return nil
	}

	defer func() {
		store.DefaultStore = nil
	}()

	return store.Close()
}

func (c *command) initSync(ctx *cli.Context) error {
	var opts []sync.Option

	if ctx.IsSet("cdm_sync_backend_username") && ctx.IsSet("cdm_sync_backend_password") {
		opts = append(opts, sync.Auth(ctx.String("cdm_store_username"), ctx.String("cdm_store_password")))
	}

	if ctx.IsSet("cdm_sync_backend_heartbeat_interval") {
		opts = append(opts, sync.HeartbeatInterval(time.Duration(ctx.Int("cdm_store_heartbeat_interval"))*time.Second))
	}

	if ctx.IsSet("cdm_sync_ttl") {
		opts = append(opts, sync.TTL(ctx.Int("cdm_sync_ttl")))
	}

	return sync.Init(ctx.String("cdm_sync_backend_service_name"), opts...)
}

func (c *command) destroySync() error {
	if sync.DefaultSync == nil {
		return nil
	}

	defer func() {
		sync.DefaultSync = nil
	}()

	return sync.Close()
}

func (c *command) initDatabase(ctx *cli.Context) error {
	var opts []database.Option

	opts = append(opts, database.SSLEnable(ctx.Bool("cdm_database_ssl_enable")))

	if ctx.IsSet("cdm_database_ssl_ca") {
		opts = append(opts, database.SSLCACert(ctx.String("cdm_database_ssl_ca")))
	}

	if ctx.IsSet("cdm_database_heartbeat_interval") {
		opts = append(opts, database.HeartbeatInterval(time.Duration(ctx.Int("cdm_database_heartbeat_interval"))*time.Second))
	}

	if ctx.IsSet("cdm_database_reconnect_interval") {
		opts = append(opts, database.ReconnectInterval(time.Duration(ctx.Int("cdm_database_reconnect_interval"))*time.Second))
	}

	// initialize default database connection config
	database.Init(
		ctx.String("cdm_database_service_name"),
		ctx.String("cdm_database_name"),
		ctx.String("cdm_database_username"),
		ctx.String("cdm_database_password"),
		opts...,
	)

	// default database connection
	return database.Connect()
}

func (c *command) destroyDatabase() error {
	return database.Close()
}

func (c *command) initLogger(ctx *cli.Context) error {
	hostname, _ := os.Hostname()
	logf, err := rotatelogs.New(
		"/var/log/"+hostname+"-%Y-%m-%d.log",
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(24*time.Hour*30),
		// 사이즈 제한: 50000000Byte = 47.68MB. 사이즈가 넘으면 [파일이름.log.0n] 으로 남음
		rotatelogs.WithRotationSize(50000000),
	)
	if err != nil {
		logger.Warnf("Unknown error : %v", err)
	}

	// initialize default cdm logger
	if err := logger.Init(
		logger.WithServiceName(ctx.App.Name),
		mLogger.WithLevel(mLogger.InfoLevel),
		mLogger.WithOutput(io.MultiWriter(os.Stdout, logf)),
		mLogger.WithCallerSkipCount(2),
	); err != nil {
		return err
	}

	// initialize watcher
	if err := watcher.Watch(ctx.App.Name); err != nil {
		return err
	}

	// initialize default micro framework logger
	var opts = []mLogger.Option{
		mLogger.WithOutput(io.MultiWriter(mLogger.DefaultLogger.Options().Out, logf)),
		mLogger.WithCallerSkipCount(mLogger.DefaultLogger.Options().CallerSkipCount),
	}

	if lv, err := mLogger.GetLevel(ctx.String("micro_logger_level")); err != nil {
		opts = append(opts, mLogger.WithLevel(mLogger.DefaultLogger.Options().Level))
	} else {
		opts = append(opts, mLogger.WithLevel(lv))
	}

	mLogger.DefaultLogger = logger.NewLogger(opts...)

	return nil
}

func (c *command) destroyLogger() error {
	return watcher.Stop()
}

func (c *command) Before(ctx *cli.Context) error {
	logger.Info("Initializing CDM Cloud framework.")
	if err := c.initDefaults(ctx); err != nil {
		logger.Errorf("Could not initialize default setting. cause:%v", err)
		return err
	}

	logger.Info("Initializing message broker.")
	if err := c.initBroker(ctx); err != nil {
		logger.Errorf("Could not initialize message broker. cause: %v", err)
		return err
	}

	logger.Info("Initializing key-value store.")
	if err := c.initStore(ctx); err != nil {
		logger.Errorf("Could not initialize key-value store. cause: %v", err)
		return err
	}

	logger.Info("Initializing sync.")
	if err := c.initSync(ctx); err != nil {
		logger.Errorf("Could not initialize sync. cause: %v", err)
		return err
	}

	logger.Info("Initializing database.")
	if err := c.initDatabase(ctx); err != nil {
		logger.Errorf("Could not initialize database connection config. cause: %v", err)
		return err
	}

	logger.Info("Initializing logger.")
	if err := c.initLogger(ctx); err != nil {
		logger.Errorf("Could not initialize logger. cause: %v", err)
		return err
	}

	logger.Info("CDM Cloud framework initialized.")
	return nil
}

func (c *command) Destroy() {
	logger.Info("Stopping logger")
	if err := c.destroyLogger(); err != nil {
		logger.Warnf("Could not stop logger. cause: %v", err)
	}

	logger.Info("Stopping database")
	if err := c.destroyDatabase(); err != nil {
		logger.Warnf("Could not stop database. cause: %v", err)
	}

	logger.Info("Closing store")
	if err := c.destroyStore(); err != nil {
		logger.Warnf("Could not close store. cause: %v", err)
	}

	logger.Info("Closing sync")
	if err := c.destroySync(); err != nil {
		logger.Warnf("Could not close sync. cause: %v", err)
	}

	logger.Info("Disconnecting broker")
	if err := c.destroyBroker(); err != nil {
		logger.Warnf("Could not disconnect broker. cause: %v", err)
	}
}

func newCommand(opts ...Option) cmd.Cmd {
	c := command{
		Cmd: cmd.NewCmd(),
		opts: Options{
			Flags:            defaultFlags,
			PersistentQueues: defaultPersistentQueues,
		},
	}

	for _, o := range opts {
		o(&c.opts)
	}

	c.App().Flags = c.opts.Flags
	c.App().Before = c.Before

	return c
}
