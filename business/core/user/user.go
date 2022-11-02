// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want to audit or something that isn't specific to the data/store layer.
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mohammadhsn/ultimate-service/business/data/store/user"
	"go.uber.org/zap"
)

// Core manages the set of APIs for user access.
type Core struct {
	log  *zap.SugaredLogger
	user user.Store
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:  log,
		user: user.NewStore(log, db),
	}
}

// Create inserts a new user into the database.
func (c Core) Create(ctx context.Context, nu user.NewUser, now time.Time) (user.User, error) {
	usr, err := c.user.Create(ctx, nu, now)

	// perform pre business operations

	if err != nil {
		return user.User{}, fmt.Errorf("create: %w", err)
	}

	// perform post business operations

	return usr, nil
}

func (c Core) Update(ctx context.Context, userId string, uu user.UpdateUser, now time.Time) error {

	if err := c.user.Update(ctx, userId, uu, now); err != nil {
		return fmt.Errorf("update %w", err)
	}

	return nil
}

func (c Core) Delete(ctx context.Context, userId string) error {
	if err := c.user.Delete(ctx, userId); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]user.User, error) {
	users, err := c.user.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query %w", err)
	}
	return users, nil
}

func (c Core) QueryById(ctx context.Context, userId string) (user.User, error) {
	usr, err := c.user.QueryById(ctx, userId)
	if err != nil {
		return user.User{}, fmt.Errorf("query: %w", err)
	}
	return usr, nil
}

func (c Core) Authenticate(ctx context.Context, now time.Time, email, password string) error {
	if err := c.user.Authenticate(ctx, now, email, password); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	return nil
}
