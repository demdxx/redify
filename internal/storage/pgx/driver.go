package pgx

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"github.com/demdxx/redify/internal/context/ctxlogger"
	"github.com/demdxx/redify/internal/keypattern"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/sql"
)

type Driver struct {
	pool   *pgxpool.Pool
	binds  []*Bind
	syntax sql.Syntax
}

func Open(ctx context.Context, connURL string) (storage.Driver, error) {
	if strings.HasPrefix(connURL, "pgx://") {
		connURL = "postgres://" + connURL[6:]
	}
	conf, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, err
	}
	logger := zapadapter.NewLogger(ctxlogger.Get(ctx).With(zap.String("driver", "pgx")))
	conf.BeforeConnect = func(ctx context.Context, conf *pgx.ConnConfig) error {
		conf.Logger = logger
		return nil
	}
	pool, err := pgxpool.ConnectConfig(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &Driver{pool: pool, syntax: sql.NewAbstractSyntax(`"`)}, nil
}

func (pg *Driver) Get(ctx context.Context, dbnum int, key string) ([]byte, error) {
	ectx := keypattern.ExecContext{}
	bind, err := pg.bindByKey(key, dbnum, ectx)
	if err != nil {
		return nil, err
	}
	rec, err := bind.Get(ctx, ectx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(rec)
}

func (pg *Driver) Set(ctx context.Context, dbnum int, key string, value []byte) error {
	ectx := keypattern.ExecContext{}
	bind, err := pg.bindByKey(key, dbnum, ectx)
	if err != nil {
		return err
	}
	return bind.Upsert(ctx, ectx, value)
}

func (pg *Driver) Del(ctx context.Context, dbnum int, key string) error {
	ectx := keypattern.ExecContext{}
	bind, err := pg.bindByKey(key, dbnum, ectx)
	if err != nil {
		return err
	}
	return bind.Del(ctx, ectx)
}

func (pg *Driver) Keys(ctx context.Context, dbnum int, pattern string) ([]string, error) {
	var (
		keys   []string
		hasKey bool
	)
	for _, bind := range pg.binds {
		ectx := keypattern.ExecContext{}
		if bind.DBNum != dbnum || !bind.MatchPattern(pattern, ectx) {
			continue
		}
		hasKey = true
		res, err := bind.List(ctx, ectx)
		if err != nil {
			return nil, err
		}
		if keys == nil {
			keys = make([]string, 0, len(res))
		}
		for _, r := range res {
			keys = append(keys, bind.Pattern.Format(r))
		}
	}
	if !hasKey {
		return nil, storage.ErrNoKey
	}
	return keys, nil
}

func (pg *Driver) Bind(ctx context.Context, conf *storage.BindConfig) error {
	var bind *Bind
	if conf.GetQuery != "" {
		bind = NewBind(pg.pool, conf.DBNum, pg.syntax,
			conf.Pattern, conf.GetQuery, conf.ListQuery, conf.UpsertQuery, conf.DelQuery)
	} else if conf.TableName != "" {
		bind = NewBindFromTableName(pg.pool, conf.DBNum, pg.syntax,
			conf.Pattern, conf.TableName, conf.WhereExt, conf.Readonly)
	} else {
		return storage.ErrInvalidBindConfig
	}
	pg.binds = append(pg.binds, bind)
	return nil
}

func (pg *Driver) Close() error {
	pg.pool.Close()
	return nil
}

func (pg *Driver) bindByKey(key string, dbnum int, ectx keypattern.ExecContext) (*Bind, error) {
	for _, b := range pg.binds {
		if b.DBNum == dbnum && b.MatchKey(key, ectx) {
			return b, nil
		}
	}
	return nil, storage.ErrNoKey
}

// SQL Example:
// CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$
//
//	DECLARE
//	    data json;
//	    notification json;
//
//	BEGIN
//
//	    -- Convert the old or new row to JSON, based on the kind of action.
//	    -- Action = DELETE?             -> OLD row
//	    -- Action = INSERT or UPDATE?   -> NEW row
//	    IF (TG_OP = 'DELETE') THEN
//	        data = row_to_json(OLD);
//	    ELSE
//	        data = row_to_json(NEW);
//	    END IF;
//
//	    -- Contruct the notification as a JSON string.
//	    notification = json_build_object(
//	                      'table',TG_TABLE_NAME,
//	                      'action', TG_OP,
//	                      'data', data);
//
//	    -- Execute pg_notify(channel, notification)
//	    PERFORM pg_notify('redify_update', notification::text);
//
//	    -- Result is ignored since this is an AFTER trigger
//	    RETURN NULL;
//	END;
//
// $$ LANGUAGE plpgsql;
//
// CREATE TRIGGER products_notify_event
// AFTER INSERT OR UPDATE OR DELETE ON products
//
//	FOR EACH ROW EXECUTE PROCEDURE notify_event();
func (pg *Driver) ListenUpdateNotifies(ctx context.Context, chanelName string, notifyFnk func(ctx context.Context, key string)) error {
	conn, err := pg.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	ctxlogger.Get(ctx).Info(`start listen notifications "` + chanelName + `"`)
	_, err = conn.Exec(context.Background(), "listen "+chanelName)
	if err != nil {
		return err
	}

	for !conn.Conn().IsClosed() {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			ctxlogger.Get(ctx).Error("pgx notification listen", zap.Error(err))
			continue
		}

		var notifyData Notification
		if err := notifyData.unmarshal([]byte(notification.Payload)); err != nil {
			ctxlogger.Get(ctx).Error("unmarshal notification message", zap.Error(err))
			continue
		}

		if ectx, err := notifyData.ectx(); err != nil {
			ctxlogger.Get(ctx).Error("unmarshal notification payload", zap.Error(err))
		} else {
			key, err := pg.keyFromTableAndContext(notifyData.Table, ectx)
			if err != nil {
				ctxlogger.Get(ctx).Error("detect key from notification", zap.Error(err))
			} else {
				notifyFnk(ctx, key)
			}
		}
	}

	ctxlogger.Get(ctx).Info(`stop listen notifications "redify_update"`)
	return nil
}

func (pg *Driver) keyFromTableAndContext(tableName string, ectx keypattern.ExecContext) (string, error) {
	for _, b := range pg.binds {
		if b.TableName() == tableName {
			return b.Pattern.Format(ectx), nil
		}
	}
	return "", storage.ErrNoKey
}

var _ storage.Driver = (*Driver)(nil)
