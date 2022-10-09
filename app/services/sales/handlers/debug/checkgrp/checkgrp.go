// Package checkgrp maintains the group of handlers for health checking.
package checkgrp

import (
	"go.uber.org/zap"
	"net/http"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}

// Readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {}

// Liveness returns simple status info if the service is alive. If the
// app is deployed to a k8s cluster, it will also return pod, node, and
// namespace details via the Downward API. The k8s environment variables
// need to be set within your Pod/Deployment manifest.
func (h Handlers) Liveness(w http.ResponseWriter, r *http.Request) {}
