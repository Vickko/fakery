package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"fakery/internal/conf"
	"fakery/internal/server"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

// getConfigPath 根据环境变量和运行环境确定配置文件路径
func getConfigPath() string {
	configDir := flagconf

	// 如果是Docker环境（通过环境变量ENVIRONMENT=docker判断）
	environment := os.Getenv("ENVIRONMENT")

	var configFile string
	if environment == "docker" {
		configFile = "config.docker.yaml"
	} else {
		configFile = "config.yaml"
	}

	// 如果configDir是文件夹，则拼接配置文件名
	if stat, err := os.Stat(configDir); err == nil && stat.IsDir() {
		return filepath.Join(configDir, configFile)
	}

	// 如果configDir直接是文件路径，则直接返回
	return configDir
}

func getLogLevel() log.Level {

	envLogLevel := os.Getenv("LOG_LEVEL")

	var logLevel log.Level
	switch strings.ToLower(envLogLevel) {
	case "debug":
		logLevel = log.LevelDebug
	case "info":
		logLevel = log.LevelInfo
	case "warn", "warning":
		logLevel = log.LevelWarn
	case "error":
		logLevel = log.LevelError
	case "fatal":
		logLevel = log.LevelFatal
	default:
		logLevel = log.LevelInfo
	}
	return logLevel
}

// createLogger 创建配置好的logger
func createLogger(logLevel log.Level) log.Logger {

	return log.With(
		log.NewFilter(
			log.NewStdLogger(os.Stdout),
			log.FilterLevel(logLevel),
		),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, registry *server.Registry) *kratos.App {
	app := kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.AfterStart(func(ctx context.Context) error {
			return registry.RegisterServers(ctx, hs, gs)
		}),
		kratos.BeforeStop(func(ctx context.Context) error {
			return registry.UnregisterServers(ctx, hs, gs)
			// return nil
		}),
	)
	return app
}

func main() {
	flag.Parse()

	logLevel := getLogLevel()
	logger := createLogger(logLevel)
	logHelper := log.NewHelper(logger)
	logHelper.Infof("Logger initialized with level: %s", logLevel.String())

	// 获取配置文件路径
	configPath := getConfigPath()
	logHelper.Infof("Loading config from: %s", configPath)

	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
