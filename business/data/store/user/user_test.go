package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/google/go-cmp/cmp"
	"github.com/mohammadhsn/ultimate-service/business/data/store/user"
	"github.com/mohammadhsn/ultimate-service/business/data/tests"
	"github.com/mohammadhsn/ultimate-service/business/sys/database"
)

var dbc = tests.DBContainer{
	Image: "postgres:14.5",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

func TestUser(t *testing.T) {
	log, db, teardown := tests.NewUnit(t, dbc)
	t.Cleanup(teardown)

	store := user.NewStore(log, db)

	t.Log("given the need to work with User records.")

	testId := 0
	t.Logf("\tTest %d:\tWhen handling a single User.", testId)
	{
		ctx := context.Background()
		now := time.Now()

		nu := user.NewUser{
			Name:            "John Doe",
			Email:           "john@doe.com",
			Roles:           []string{},
			Password:        "gophers",
			PasswordConfirm: "gophers",
		}

		usr, err := store.Create(ctx, nu, now)
		if err != nil {
			t.Fatalf("\t%s\tTest %d: \tShould be able to create user: %s.", tests.Failed, testId, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to create user.", tests.Success, testId)

		saved, err := store.QueryById(ctx, usr.ID)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", tests.Failed, testId, err)
		}

		t.Logf("\t%s\tTest %d:\tShould be able to retrieve user.", tests.Success, testId)

		//if diff := cmp.Diff(usr, saved); diff != "" {
		//	t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", tests.Failed, testId, diff)
		//}
		//t.Logf("\t%s\tTest %d:\tShould be able to get back user.", tests.Success, testId)

		upd := user.UpdateUser{
			Name:  tests.StringPointer("john foo"),
			Email: tests.StringPointer("foo@bar.com"),
		}

		if err := store.Update(ctx, usr.ID, upd, now); err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to update user by ID: %s.", tests.Failed, testId, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to update user.", tests.Success, testId)

		saved, err = store.QueryByEmail(ctx, *upd.Email)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by email: %s.", tests.Failed, testId, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by email.", tests.Success, testId)

		if saved.Name != *upd.Name {
			t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Name: %s.", tests.Failed, testId, err)
			t.Logf("\t\tTest %d:\tGot: %s", testId, saved.Name)
			t.Logf("\t\tTest %d:\tExp: %s", testId, *upd.Name)
		} else {
			t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name: %s.", tests.Success, testId, err)
		}

		if saved.Email != *upd.Email {
			t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Email: %s.", tests.Failed, testId, err)
			t.Logf("\t\tTest %d:\tGot: %v", testId, saved.Email)
			t.Logf("\t\tTest %d:\tExp: %v", testId, *upd.Email)
		} else {
			t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email: %s.", tests.Success, testId, err)
		}

		if err := store.Delete(ctx, usr.ID); err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to delete user: %s.", tests.Failed, testId, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to delete user: %s.", tests.Success, testId, err)

		_, err = store.QueryById(ctx, usr.ID)
		if !errors.Is(err, database.ErrNotFound) {
			t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", tests.Failed, testId, err)
		}
		t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", tests.Failed, testId, err)
	}
}
