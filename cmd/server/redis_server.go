package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/redcon"
	"go.uber.org/zap"

	"github.com/demdxx/redify/internal/context/ctxlogger"
	"github.com/demdxx/redify/internal/storage"
)

type userContext struct {
	DBNum int
}

// RedisServer redify implement basic functionality of the redis proxy
type RedisServer struct {
	Driver         storage.Driver
	RequestTimeout time.Duration

	ctx context.Context
	ps  redcon.PubSub
}

// ListenAndServe redify RedisServer
func (srv *RedisServer) ListenAndServe(ctx context.Context, addr string) error {
	srv.ctx = ctx
	return redcon.ListenAndServe(addr,
		srv.command,
		srv.acceptConnection,
		srv.closeConnection)
}

func (srv *RedisServer) command(conn redcon.Conn, cmd redcon.Command) {
	ctx := srv.context()
	if srv.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(srv.context(), srv.RequestTimeout)
		defer cancel()
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		ctxlogger.Get(ctx).Debug("command",
			zap.String("cmd", string(cmd.Args[0])),
			zap.Int("args", len(cmd.Args)-1),
			zap.Duration("time", duration))
	}()

	rCtx := getUserContext(conn.Context())
	dbnum := rCtx.DBNum
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
	case "publish":
		// Publish to all pub/sub subscribers and return the number of
		// messages that were sent.
		if len(cmd.Args) != 3 {
			srv.wrongNumberArgsError(conn, cmd)
			return
		}
		count := srv.ps.Publish(string(cmd.Args[1]), string(cmd.Args[2]))
		conn.WriteInt(count)
	case "subscribe", "psubscribe":
		// Subscribe to a pub/sub channel. The `Psubscribe` and
		// `Subscribe` operations will detach the connection from the
		// event handler and manage all network I/O for this connection
		// in the background.
		if len(cmd.Args) < 2 {
			srv.wrongNumberArgsError(conn, cmd)
			return
		}
		command := strings.ToLower(string(cmd.Args[0]))
		for i := 1; i < len(cmd.Args); i++ {
			if command == "psubscribe" {
				srv.ps.Psubscribe(conn, string(cmd.Args[i]))
			} else {
				srv.ps.Subscribe(conn, string(cmd.Args[i]))
			}
		}
	case "detach":
		hconn := conn.Detach()
		log.Printf("connection has been detached")
		go func() {
			defer hconn.Close()
			hconn.WriteString("OK")
			hconn.Flush()
		}()
	case "ping":
		conn.WriteString("PONG")
	case "quit":
		conn.WriteString("OK")
		conn.Close()
	case "set":
		srv.cmdSet(ctx, conn, dbnum, cmd)
	case "mset":
		srv.cmdMSet(ctx, conn, dbnum, cmd)
	case "get":
		srv.cmdGet(ctx, conn, dbnum, cmd)
	case "mget":
		srv.cmdMGet(ctx, conn, dbnum, cmd)
	case "hget":
		srv.cmdHGet(ctx, conn, dbnum, cmd)
	case "hgetall":
		srv.cmdHGetall(ctx, conn, dbnum, cmd)
	case "del":
		srv.cmdDel(ctx, conn, dbnum, cmd)
	case "keys":
		srv.cmdKeys(ctx, conn, dbnum, cmd)
	case "select":
		if len(cmd.Args) != 2 {
			srv.wrongNumberArgsError(conn, cmd)
			return
		}
		dbnum, _ = strconv.Atoi(string(cmd.Args[1]))
		if dbnum < 0 || dbnum > 15 {
			conn.WriteError("ERR invalid database number " + strconv.Itoa(dbnum) + ", must be from 0 to 15")
			return
		}
		rCtx.DBNum = dbnum
		conn.SetContext(rCtx)
		conn.WriteString("OK")
	case "config":
		conn.WriteArray(2)
		conn.WriteBulk(cmd.Args[2])
		conn.WriteBulkString("")
	}
}

func (srv *RedisServer) acceptConnection(conn redcon.Conn) bool {
	return true
}

func (srv *RedisServer) closeConnection(conn redcon.Conn, err error) {
}

func (srv *RedisServer) cmdGet(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		key        = string(cmd.Args[1])
		value, err = srv.Driver.Get(ctx, dbnum, key)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		conn.WriteError("ERR " + err.Error())
		return
	}
	if value == nil {
		conn.WriteNull()
	} else {
		conn.WriteBulk(value)
	}
}

func (srv *RedisServer) cmdMGet(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) < 2 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	conn.WriteArray(len(cmd.Args) - 1)
	for i := 1; i < len(cmd.Args); i++ {
		var (
			key        = string(cmd.Args[i])
			value, err = srv.Driver.Get(ctx, dbnum, key)
		)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			ctxlogger.Get(ctx).Error("mget value", zap.Error(err), zap.String("key", key))
		}
		if value == nil {
			conn.WriteNull()
		} else {
			conn.WriteBulk(value)
		}
	}
}

func (srv *RedisServer) cmdHGet(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 3 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		key        = string(cmd.Args[1])
		name       = string(cmd.Args[2])
		value, err = srv.Driver.Get(ctx, dbnum, key)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		ctxlogger.Get(ctx).Error("mget value", zap.Error(err), zap.String("key", key))
	}
	if value == nil {
		conn.WriteNull()
	} else if bytes.HasPrefix(value, []byte("{")) {
		var record map[string]json.RawMessage
		if err := json.Unmarshal(value, &record); err == nil {
			if val := record[name]; bytes.HasPrefix(val, []byte(`"`)) {
				var s string
				if err := json.Unmarshal(val, &s); err != nil {
					conn.WriteError("ERR decode field '" + key + "' " + err.Error())
				} else {
					conn.WriteBulkString(s)
				}
			} else {
				conn.WriteBulk(val)
			}
		} else {
			conn.WriteNull()
		}
	} else {
		conn.WriteNull()
	}
}

func (srv *RedisServer) cmdHGetall(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		key        = string(cmd.Args[1])
		value, err = srv.Driver.Get(ctx, dbnum, key)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		ctxlogger.Get(ctx).Error("mget value", zap.Error(err), zap.String("key", key))
	}
	if value == nil {
		conn.WriteNull()
	} else if bytes.HasPrefix(value, []byte("{")) {
		var record map[string]json.RawMessage
		if err := json.Unmarshal(value, &record); err == nil {
			conn.WriteArray(len(record) * 2)
			for key, val := range record {
				conn.WriteBulkString(key)
				if bytes.HasPrefix(val, []byte(`"`)) {
					var s string
					if err := json.Unmarshal(val, &s); err != nil {
						conn.WriteError("ERR decode field '" + key + "' " + err.Error())
					} else {
						conn.WriteBulkString(s)
					}
				} else {
					conn.WriteBulk(val)
				}
			}
		} else {
			conn.WriteArray(1)
			conn.WriteBulk(value)
		}
	} else {
		conn.WriteBulk(value)
	}
}

func (srv *RedisServer) cmdSet(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 3 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		key   = string(cmd.Args[1])
		value = cmd.Args[2]
		err   = srv.Driver.Set(ctx, dbnum, key, value)
	)
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}
	conn.WriteString("OK")
}

func (srv *RedisServer) cmdMSet(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if (len(cmd.Args)-1)%2 != 0 || len(cmd.Args) < 3 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	for i := 0; i < len(cmd.Args); i += 2 {
		var (
			key   = string(cmd.Args[i])
			value = cmd.Args[i+1]
			err   = srv.Driver.Set(ctx, dbnum, key, value)
		)
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return
		}
	}
	conn.WriteString("OK")
}

func (srv *RedisServer) cmdDel(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		key = string(cmd.Args[1])
		err = srv.Driver.Del(ctx, dbnum, key)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		conn.WriteError("ERR " + err.Error())
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		conn.WriteInt(0)
	} else {
		conn.WriteInt(1)
	}
}

func (srv *RedisServer) cmdKeys(ctx context.Context, conn redcon.Conn, dbnum int, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		srv.wrongNumberArgsError(conn, cmd)
		return
	}
	var (
		pattern   = string(cmd.Args[1])
		keys, err = srv.Driver.Keys(ctx, dbnum, pattern)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		conn.WriteError("ERR " + err.Error())
		return
	}
	conn.WriteArray(len(keys))
	for _, key := range keys {
		conn.WriteBulkString(key)
	}
}

func (srv *RedisServer) wrongNumberArgsError(conn redcon.Conn, cmd redcon.Command) {
	conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
}

func (srv *RedisServer) context() context.Context {
	if srv.ctx != nil {
		return srv.ctx
	}
	return context.Background()
}

func getUserContext(ctx any) *userContext {
	switch v := ctx.(type) {
	case *userContext:
		return v
	case nil:
		return &userContext{}
	default:
		return &userContext{}
	}
}
