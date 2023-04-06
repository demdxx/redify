package sql

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestBind1(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxdb := sqlx.NewDb(db, "test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var (
		bind = NewBind(
			sqlxdb, 0,
			NewAbstractSyntax(`"`),
			"users_{{username}}",
			"SELECT * FROM users WHERE username={{username}}",
			"SELECT username FROM users",
			"INSERT INTO users (username) VALUES({{username}})",
			"DELETE FROM users WHERE username={{username}}",
			nil,
			false,
		)
		ectx = keypattern.ExecContext{
			"username": "testuser",
		}
	)
	testBindCommon(ctx, t, mock, bind, ectx)

	mock.ExpectClose()
	assert.NoError(t, db.Close())
}

func TestBind2(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxdb := sqlx.NewDb(db, "test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var (
		bind = NewBindFromTableName(
			sqlxdb, 0,
			NewAbstractSyntax(`"`),
			"users_{{username}}",
			"users", "",
			false, nil, false,
		)
		ectx = keypattern.ExecContext{
			"username": "testuser",
		}
	)
	testBindCommon(ctx, t, mock, bind, ectx)

	mock.ExpectClose()
	assert.NoError(t, db.Close())
}

func testBindCommon(ctx context.Context, t *testing.T, mock sqlmock.Sqlmock, bind *Bind, ectx keypattern.ExecContext) {
	t.Run("negative key", func(t *testing.T) {
		if bind.MatchKey("not_users_key", ectx) {
			t.Error("invalid negative key matching")
		}
	})

	t.Run("select record", func(t *testing.T) {
		columns := []string{"id", "username"}
		mock.ExpectQuery("SELECT \\*").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(
				sqlmock.NewRows(columns).
					AddRow(1, "testuser"),
			)

		if !bind.MatchKey("users_1", ectx) {
			t.Error("invalid key matching")
		}

		rec, err := bind.Get(ctx, ectx)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, gocast.Int(rec["id"]))
			assert.Equal(t, "testuser", rec["username"])
		}
	})
	t.Run("select list", func(t *testing.T) {
		columns := []string{"id", "username"}
		mock.ExpectQuery("SELECT").
			WillReturnRows(
				sqlmock.NewRows(columns).
					AddRow(1, "testuser1").
					AddRow(2, "testuser2"),
			)
		if !bind.MatchPattern("users_*", ectx) {
			t.Error("invalid key matching")
		}
		res, err := bind.List(ctx, ectx)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(res))
			assert.Equal(t, 1, gocast.Int(res[0]["id"]))
			assert.Equal(t, 2, gocast.Int(res[1]["id"]))
		}
	})
	t.Run("insert record", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		ectx["username"] = "testuser"
		err := bind.Upsert(ctx, ectx, []byte(`{"newVar":"val"}`))
		assert.NoError(t, err)
	})
	t.Run("delete record", func(t *testing.T) {
		mock.ExpectExec("DELETE").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		ectx["username"] = "testuser"
		err := bind.Del(ctx, ectx)
		assert.NoError(t, err)
	})
	t.Run("readnly", func(t *testing.T) {
		bind.DelQuery = nil
		bind.UpsertQuery = nil
		err := bind.Upsert(ctx, ectx, []byte(`{}`))
		assert.ErrorIs(t, err, storage.ErrReadOnly)
		err = bind.Del(ctx, ectx)
		assert.ErrorIs(t, err, storage.ErrReadOnly)
	})
}
