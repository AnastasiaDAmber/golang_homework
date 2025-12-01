package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/app"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := NewConfigFromFile(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logg := logger.New(cfg.Logger.Level)

	var storage app.Storage
	switch cfg.Storage.Type {
	case "sql":
		sql := sqlstorage.New(cfg.DB.DSN)
		if err := sql.Connect(context.Background()); err != nil {
			logg.Error("failed to connect to db: " + err.Error())
			os.Exit(1)
		}

		storage = sql
	default:
		storage = memorystorage.New()
	}

	calendar := app.New(logg, storage)

	httpServer := internalhttp.NewServer(logg, calendar, cfg.Server.Host, cfg.Server.HTTPPort)
	grpcServer := internalgrpc.NewServer(logg, calendar, cfg.Server.Host, cfg.Server.GRPCPort)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpServer.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
		if err := grpcServer.Stop(ctx); err != nil {
			logg.Error("failed to stop grpc server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	// Запускаем оба сервера в отдельных горутинах
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
			cancel()
		}
	}()

	// HTTP сервер запускаем в основной горутине
	if err := httpServer.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
