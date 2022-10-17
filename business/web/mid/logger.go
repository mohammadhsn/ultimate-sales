package mid

import (
	"context"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Logger(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is mussing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return err
			}

			log.Infow(
				"request started",
				"traceID", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteAddr", r.RemoteAddr,
			)

			// Call the next handler
			err = handler(ctx, w, r)

			log.Infow(
				"request completed",
				"traceID", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteAddr", r.RemoteAddr,
				"statusCode", v.StatusCode,
				"since", time.Since(v.Now),
			)

			// Return the error, so it can be handled further up the chain.
			return err
		}
	}

	return m
}
