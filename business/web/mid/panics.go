package mid

import (
	"context"
	"fmt"
	"github.com/mohammadhsn/ultimate-service/business/sys/metrics"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"net/http"
	"runtime/debug"
)

func Panics() web.Middleware {
	return func(handler web.Handler) web.Handler {
		// Create the handler that will be attached in the middleware chain.
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					// Stack trace will be provided.
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(debug.Stack()))

					// Updates the metrics stored in the context.
					metrics.AddPanics(ctx)
				}
			}()
			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}
	}
}
