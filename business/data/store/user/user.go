package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mohammadhsn/ultimate-service/business/sys/database"
	"github.com/mohammadhsn/ultimate-service/business/sys/validate"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Store manages the set of APIs for user access.
type Store struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

// NewStore constructs a user store for api access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	if err := validate.Check(nu); err != nil {
		return User{}, fmt.Errorf("validating data: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)

	if err != nil {
		return User{}, fmt.Errorf("generating password hash: %w", err)
	}

	usr := User{
		ID:           validate.GenerateId(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now,
		DateUpdated:  now,
	}

	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
	    (:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return User{}, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

func (s Store) Delete(ctx context.Context, userId string) error {
	if err := validate.CheckId(userId); err != nil {
		return database.ErrInvalidID
	}

	data := struct {
		UserId string `db:"user_id"`
	}{
		UserId: userId,
	}

	const q string = `DELETE FROM users WHERE user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting userId[%s]: %w", userId, err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, userId string, uu UpdateUser, now time.Time) error {
	if err := validate.CheckId(userId); err != nil {
		return database.ErrInvalidID
	}

	if err := validate.Check(uu); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	usr, err := s.QueryById(ctx, userId)

	if err != nil {
		return fmt.Errorf("updating user userId[%s]: %w", userId, err)
	}
	if uu.Name != nil {
		usr.Name = *uu.Name
	}
	if uu.Email != nil {
		usr.Email = *uu.Email
	}
	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}
	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating password hash: %w", err)
		}
		usr.PasswordHash = pw
	}
	usr.DateUpdated = now

	const q string = `
	UPDATE
		users
	SET
	    "name" = :name,
	    "email" = :email,
	    "roles" = :roles,
	    "password_hash" = :password_hash,
	    "date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return fmt.Errorf("updating userId[%s]: %w", userId, err)
	}

	return nil
}

func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q string = `SELECT * FROM users ORDER BY user_id OFFSET :offset ROWS FETCH NEXT :rows_per_page ONLY`

	var users []User
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &users); err != nil {
		if err == database.ErrNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return users, nil
}

// QueryById gets the specified user from the database.
func (s Store) QueryById(ctx context.Context, userId string) (User, error) {
	if err := validate.CheckId(userId); err != nil {
		return User{}, database.ErrInvalidID
	}

	data := struct {
		UserId string `db:"user_id"`
	}{
		UserId: userId,
	}

	const q string = `SELECT * FROM users WHERE user_id = :user_id`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return User{}, database.ErrNotFound
		}
	}

	return usr, nil
}

func (s Store) QueryByEmail(ctx context.Context, email string) (User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q string = `SELECT * FROM users WHERE email=:email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if err == database.ErrNotFound {
			return User{}, database.ErrNotFound
		}
		return User{}, fmt.Errorf("selecting email[%q]: %w", email, err)
	}

	return usr, nil
}

func (s Store) Authenticate(ctx context.Context, now time.Time, email, password string) error {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q string = `SELECT * FROM users WHERE email = :email`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, usr); err != nil {
		if err == database.ErrNotFound {
			return err
		}
		return fmt.Errorf("selecting user[%q]: %w", email, err)
	}

	// Compare the provided password with the saved hash. User the bcrypt
	// comparison function, so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return database.ErrAtuthenticationFailure
	}

	// If we are this far the request is valid. Create some claim for the user
	// and generate their token.
	// Todo: Create auth.Claims type
	return nil
}
