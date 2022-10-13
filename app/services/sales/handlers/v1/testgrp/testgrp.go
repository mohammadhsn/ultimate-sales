package testgrp

import (
	"context"
	"encoding/json"
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
	json.NewEncoder(w).Encode(status)

	statusCode := http.StatusOK

	h.Log.Infow(
		"liveness",
		"statusCode", statusCode,
		"method", r.Method,
		"path", r.URL.Path,
		"remoteAddr", r.RemoteAddr,
	)

	return nil
}
