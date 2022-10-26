package handlers

import (
	"expvar"
	"github.com/jmoiron/sqlx"
	"github.com/mohammadhsn/ultimate-service/app/services/sales/handlers/debug/checkgrp"
	"github.com/mohammadhsn/ultimate-service/app/services/sales/handlers/v1/testgrp"
	"github.com/mohammadhsn/ultimate-service/business/web/mid"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"net/http"
	"net/http/pprof"
	"os"

	"go.uber.org/zap"
)

func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Symbol)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library and then custom
// debug application routes for the service. This bypassing the yse of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}

// APIMuxConfig constructs a http.Handler with all application routes defined.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

func APIMux(cfg APIMuxConfig) *web.App {
	// Construct the web.App which holds all routes.
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	// Load the routes for the different versions of the API.
	v1(app, cfg)

	return app
}

func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"

	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/test", tgh.Test)
}
