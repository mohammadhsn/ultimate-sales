package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mohammadhsn/ultimate-service/foundation/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	_ "github.com/lib/pq" // for the init() purpose
)

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Config is the required properties to use the database.
type Config struct {
	User        string
	Password    string
	Host        string
	Name        string
	MaxIdleCons int
	MaxOpenCons int
	DisableTLS  bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)

	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleCons)
	db.SetMaxOpenConns(cfg.MaxOpenCons)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't time out or be cancelled.

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to datetime connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, query string, data interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedExecContext", "traceID", web.GetTraceID(ctx), "query", q)

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "database.query")
	span.SetAttributes(attribute.String("query", q))
	defer span.End()

	if _, err := db.NamedExecContext(ctx, query, data); err != nil {
		return err
	}

	return nil
}

// NamedQuerySlice is a le;per function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func NamedQuerySlice(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedQuerySlice", "traceID", web.GetTraceID(ctx), "query", q)

	val := reflect.ValueOf(dest)

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}

	return nil
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type.
func NamedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedQueryStruct", "traceID", web.GetTraceID(ctx), "query", q)

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}
	if !rows.Next() {
		return ErrNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args ...interface{}) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}

		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
