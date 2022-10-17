package testgrp

import (
	"context"
	"errors"
	"github.com/mohammadhsn/ultimate-service/business/sys/validate"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

// Test handler is for development.
func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
