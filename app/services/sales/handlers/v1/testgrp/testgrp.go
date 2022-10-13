package testgrp

import (
	"context"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"go.uber.org/zap"
	"net/http"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

// Test handler is for development.
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	h.Log.Infow(
		"liveness",
		"statusCode", http.StatusOK,
		"method", r.Method,
		"path", r.URL.Path,
		"remoteAddr", r.RemoteAddr,
	)

	return web.Respond(ctx, w, status, http.StatusOK)
}
