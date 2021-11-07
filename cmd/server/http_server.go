package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/demdxx/redify/internal/storage"
)

type HTTPServer struct {
	Driver         storage.Driver
	RequestTimeout time.Duration
}

// ListenAndServe redify HTTPServer
func (srv *HTTPServer) ListenAndServe(ctx context.Context, addr string) error {
	srvApp := fiber.New(fiber.Config{
		ServerHeader: "redify",
		ReadTimeout:  srv.RequestTimeout,
	})

	srvApp.Get("/:dbnum/:key", srv.get)
	srvApp.Put("/:dbnum/:key", srv.set)
	srvApp.Post("/:dbnum/:key", srv.set)
	srvApp.Get("/:dbnum/list/:pattern", srv.list)
	srvApp.Delete("/:dbnum/:key", srv.del)

	return srvApp.Listen(addr)
}

func (srv *HTTPServer) get(c *fiber.Ctx) error {
	var (
		ctx        = c.UserContext()
		key        = c.Params("key")
		dbnum, _   = c.ParamsInt("dbnum")
		value, err = srv.Driver.Get(ctx, dbnum, key)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return sendError(c, err)
	}
	if value == nil {
		return sendNotFound(c)
	}
	return sendJSON(c, value)
}

func (srv *HTTPServer) list(c *fiber.Ctx) error {
	var (
		ctx       = c.UserContext()
		pattern   = c.Params("pattern")
		dbnum, _  = c.ParamsInt("dbnum")
		keys, err = srv.Driver.Keys(ctx, dbnum, pattern)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return sendError(c, err)
	}
	if keys == nil {
		return sendNotFound(c)
	}
	return sendJSONObject(c, keys)
}

func (srv *HTTPServer) set(c *fiber.Ctx) error {
	var (
		ctx      = c.UserContext()
		key      = c.Params("key")
		dbnum, _ = c.ParamsInt("dbnum")
		err      = srv.Driver.Set(ctx, dbnum, key, c.Body())
	)
	if err != nil {
		return sendError(c, err)
	}
	return sendOK(c)
}

func (srv *HTTPServer) del(c *fiber.Ctx) error {
	var (
		ctx      = c.UserContext()
		key      = c.Params("key")
		dbnum, _ = c.ParamsInt("dbnum")
		err      = srv.Driver.Del(ctx, dbnum, key)
	)
	if errors.Is(err, storage.ErrNotFound) {
		return sendNotFound(c)
	}
	if err != nil {
		return sendError(c, err)
	}
	return sendOK(c)
}

func sendNotFound(c *fiber.Ctx) error {
	c.SendStatus(fiber.StatusNotFound)
	c.Response().Header.Add("content-type", "application/json")
	return c.SendString(`{"status":"error","error":"not found"}`)
}

func sendOK(c *fiber.Ctx) error {
	return sendJSON(c, []byte(`{"status":"OK"}`))
}

func sendError(c *fiber.Ctx, err error) error {
	return sendJSON(c, []byte(`{"status":"error","error":"`+
		strings.ReplaceAll(err.Error(), `"`, `\"`)+`"}`))
}

func sendJSONObject(c *fiber.Ctx, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return sendJSON(c, data)
}

func sendJSON(c *fiber.Ctx, data []byte) error {
	c.SendStatus(fiber.StatusOK)
	c.Response().Header.Add("content-type", "application/json")
	var buf bytes.Buffer
	_, _ = buf.Write([]byte(`{"status":"OK", "result":`))
	_, _ = buf.Write(data)
	_, _ = buf.Write([]byte(`}`))
	return c.Send(buf.Bytes())
}
