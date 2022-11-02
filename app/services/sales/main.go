package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/mohammadhsn/ultimate-service/app/services/sales/handlers"
	"github.com/mohammadhsn/ultimate-service/business/sys/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	//_ "go.uber.org/automaxprocs"
)

var build = "develop"

func main() {
	// Construct the application logger.
	log, err := initLogger("SALES")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer log.Sync()
	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	// Configuration
	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimout      time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			AreYouOk        bool          `conf:"default:true"`
		}
		DB struct {
			User        string `conf:"default:postgres"`
			Password    string `conf:"default:postgres,mask"`
			Host        string `conf:"default:localhost"`
			Name        string `conf:"default:postgres"`
			MaxIdleCons int    `conf:"default:0"`
			MaxOpenCons int    `conf:"default:0"`
			DisableTLS  bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright stuff",
		},
	}

	const prefix = "SALES"
	help, err := conf.ParseOSArgs(prefix, &cfg)

	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing config %w", err)
	}

	// Application starting
	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)

	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}

	log.Infow("startup", "config", out)

	expvar.NewString("build").Set(build)

	// Database support

	// Create connectivity to the database.
	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	cfgDB := database.Config{
		User:        cfg.DB.User,
		Password:    cfg.DB.Password,
		Host:        cfg.DB.Host,
		Name:        cfg.DB.Name,
		MaxIdleCons: cfg.DB.MaxIdleCons,
		MaxOpenCons: cfg.DB.MaxOpenCons,
		DisableTLS:  cfg.DB.DisableTLS,
	}

	db, err := database.Open(cfgDB)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	log.Infow("startup", "status", "initializing API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGKILL)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimout,
		WriteTimeout: cfg.Web.ReadTimout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Info("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion,
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefuly: %w", err)
		}

	}

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()

	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
