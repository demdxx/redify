package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/driftprogramming/pgxpoolmock"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var (
		mockPool = pgxpoolmock.NewMockPgxPool(ctrl)
		bind     = NewBind(
			mockPool,
			0,
			"users_{{username}}",
			"SEELCT * FROM users WHERE username={{username}}",
			"SELECT username FROM users",
			"INSERT INTO users (username) VALUES({{username}})",
			"DELETE FROM users WHERE username={{username}}",
		)
		ectx = keypattern.ExecContext{
			"username": "testuser",
		}
	)

	t.Run("negative key", func(t *testing.T) {
		if bind.matchKey("not_users_key", ectx) {
			t.Error("invalid negative key matching")
		}
	})

	t.Run("select record", func(t *testing.T) {
		columns := []string{"id", "username"}
		pgxRows := pgxpoolmock.NewRows(columns).
			AddRow(1, "testuser").ToPgxRows()
		mockPool.EXPECT().
			Query(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf("")).
			Return(pgxRows, nil)

		if !bind.matchKey("users_1", ectx) {
			t.Error("invalid key matching")
		}

		rec, err := bind.get(ctx, ectx)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, rec["id"])
			assert.Equal(t, "testuser", rec["username"])
		}
	})
	t.Run("select list", func(t *testing.T) {
		columns := []string{"id", "username"}
		pgxRows := pgxpoolmock.NewRows(columns).
			AddRow(1, "testuser1").
			AddRow(2, "testuser2").ToPgxRows()
		mockPool.EXPECT().
			Query(gomock.Any(), gomock.Any()).
			Return(pgxRows, nil)
		if !bind.matchPattern("users_*", ectx) {
			t.Error("invalid key matching")
		}
		res, err := bind.list(ctx, ectx)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(res))
			assert.Equal(t, 1, res[0]["id"])
			assert.Equal(t, 2, res[1]["id"])
		}
	})
	t.Run("insert record", func(t *testing.T) {
		mockPool.EXPECT().
			Exec(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf("")).
			Return(pgconn.CommandTag("INSERT"), nil)
		ectx["username"] = "testuser"
		err := bind.upsert(ctx, ectx, []byte(`{"newVar":"val"}`))
		assert.NoError(t, err)
	})
	t.Run("delete record", func(t *testing.T) {
		mockPool.EXPECT().
			Exec(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf("")).
			Return(pgconn.CommandTag("DELETE"), nil)
		ectx["username"] = "testuser"
		err := bind.del(ctx, ectx)
		assert.NoError(t, err)
	})
	t.Run("readnly", func(t *testing.T) {
		bind.delQuery = nil
		bind.upsertQuery = nil
		err := bind.upsert(ctx, ectx, []byte(`{}`))
		assert.ErrorIs(t, err, storage.ErrReadOnly)
		err = bind.del(ctx, ectx)
		assert.ErrorIs(t, err, storage.ErrReadOnly)
	})
}
