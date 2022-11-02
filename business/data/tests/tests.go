// Package tests contains supporting code for running tests.
package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mohammadhsn/ultimate-service/business/data/schema"
	"github.com/mohammadhsn/ultimate-service/business/sys/database"
	"github.com/mohammadhsn/ultimate-service/foundation/docker"
	"github.com/mohammadhsn/ultimate-service/foundation/logger"
	"go.uber.org/zap"
)

const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// DBContainer provides configuration for a container to run.
type DBContainer struct {
	Image string
	Port  string
	Args  []string
}

// NewUnit creates a test database inside a Docker container. it creates the
// required table structure but the database os otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, dbc DBContainer) (*zap.SugaredLogger, *sqlx.DB, func()) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	c := docker.StartContainer(t, dbc.Image, dbc.Port, dbc.Args...)

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})

	if err != nil {
		t.Fatalf("openning database connection: %c", err)
	}

	t.Log("waiting for database to be ready ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.Id)
		docker.StopContainer(t, c.Id)
		t.Fatalf("Migrating error: %s", err)
	}

	if err := schema.Seed(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.Id)
		docker.StopContainer(t, c.Id)
		t.Fatalf("seeding error: %s", err)
	}

	log, err := logger.New("TEST")
	if err != nil {
		t.Fatalf("logger error: %s", err)
	}

	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.Id)

		log.Sync()

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old
		fmt.Println("*************************** LOGS ***************************")
		fmt.Println(buf.String())
		fmt.Println("*************************** LOGS ***************************")
	}

	return log, db, teardown
}

func StringPointer(s string) *string {
	return &s
}

func IntPointer(i int) *int {
	return &i
}
