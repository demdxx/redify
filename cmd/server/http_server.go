package server

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/xtypes"
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
	srvApp.Use(logger.New(logger.ConfigDefault))

	srvApp.Get("/:dbnum/:key", srv.get)
	srvApp.Put("/:dbnum/:key", srv.set)
	srvApp.Post("/:dbnum/:key", srv.set)
	srvApp.Get("/:dbnum/list/:pattern", srv.list)
	srvApp.Get("/:dbnum/keys/:pattern", srv.keys)
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
	if err != nil && !errors.Is(err, storage.ErrNotFound) && !errors.Is(err, storage.ErrNoKey) {
		return sendError(c, err)
	}
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, storage.ErrNoKey) {
		return sendNotFound(c)
	}
	return sendJSON(c, value)
}

func (srv *HTTPServer) keys(c *fiber.Ctx) error {
	var (
		ctx       = c.UserContext()
		pattern   = c.Params("pattern")
		dbnum, _  = c.ParamsInt("dbnum")
		keys, err = srv.Driver.Keys(ctx, dbnum, pattern)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) && !errors.Is(err, storage.ErrNoKey) {
		return sendError(c, err)
	}
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, storage.ErrNoKey) {
		return sendNotFound(c)
	}
	return sendJSONObject(c, keys)
}

func (srv *HTTPServer) list(c *fiber.Ctx) error {
	var (
		ctx       = c.UserContext()
		pattern   = c.Params("pattern")
		dbnum, _  = c.ParamsInt("dbnum")
		format    = strings.ToLower(c.Query("format", "json"))
		data, err = srv.Driver.List(ctx, dbnum, pattern)
	)
	if err != nil && !errors.Is(err, storage.ErrNotFound) && !errors.Is(err, storage.ErrNoKey) {
		return sendError(c, err)
	}
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, storage.ErrNoKey) {
		return sendNotFound(c)
	}
	return sendResponseFromat(c, format, data)
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
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, storage.ErrNoKey) {
		return sendNotFound(c)
	}
	if err != nil {
		return sendError(c, err)
	}
	return sendOK(c)
}

func sendNotFound(c *fiber.Ctx) error {
	_ = c.SendStatus(fiber.StatusNotFound)
	c.Response().Header.Add("content-type", "application/json")
	return c.SendString(`{"status":"error","error":"not found"}`)
}

func sendOK(c *fiber.Ctx) error {
	return sendJSON(c, []byte(`{"status":"OK"}`))
}

func sendError(c *fiber.Ctx, err error) error {
	_ = c.SendStatus(fiber.StatusInternalServerError)
	c.Response().Header.Add("content-type", "application/json")
	return c.SendString(`{"status":"error","error":"` +
		strings.ReplaceAll(err.Error(), `"`, `\"`) + `"}`)
}

func sendJSONObject(c *fiber.Ctx, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return sendJSON(c, data)
}

func sendJSON(c *fiber.Ctx, data []byte) error {
	_ = c.SendStatus(fiber.StatusOK)
	c.Response().Header.Add("content-type", "application/json")
	var buf bytes.Buffer
	_, _ = buf.Write([]byte(`{"status":"OK", "result":`))
	_, _ = buf.Write(data)
	_, _ = buf.Write([]byte(`}`))
	return c.Send(buf.Bytes())
}

func sendResponseFromat(c *fiber.Ctx, format string, data []storage.Record) error {
	_ = c.Status(fiber.StatusOK)
	switch format {
	case "json":
		c.Response().Header.Add("content-type", "application/json")
		enc := json.NewEncoder(c)
		_, _ = c.WriteString(`{"status":"OK","result":[`)
		for i, r := range data {
			if i > 0 {
				_, _ = c.Write([]byte(`,`))
			}
			_ = enc.Encode(r)
		}
		_, _ = c.WriteString(`]}`)
	case "jsonflat":
		c.Response().Header.Add("content-type", "application/json")
		enc := json.NewEncoder(c)
		for _, r := range data {
			_ = enc.Encode(r)
		}
	case "jsonarray", "jsonarr", "jsonlist":
		c.Response().Header.Add("content-type", "application/json")
		enc := json.NewEncoder(c)
		_, _ = c.WriteString(`[`)
		for i, r := range data {
			if i > 0 {
				_, _ = c.Write([]byte(`,`))
			}
			_ = enc.Encode(r)
		}
		_, _ = c.WriteString(`]`)
	case "csv", "csvheadless":
		c.Response().Header.Add("content-type", "text/csv")

		skipHeader := format == "csvheadless" || gocast.Bool(c.Query("skipHeader"))
		keys := xtypes.Slice[string](strings.Split(c.Query("keys"), ",")).
			Filter(func(s string) bool { return strings.TrimSpace(s) != "" })

		// Extract headers from data
		if len(keys) == 0 {
			headers := map[string]bool{}
			for _, r := range data {
				for k := range r {
					headers[k] = true
				}
			}
			keys = xtypes.Map[string, bool](headers).Keys()
		}

		// Write data
		enc := csv.NewWriter(c)
		if !skipHeader {
			_ = enc.Write(keys)
		}
		row := make([]string, 0, len(keys))
		for _, r := range data {
			row = row[:0]
			for _, k := range keys {
				row = append(row, r.GetString(k))
			}
			_ = enc.Write(row)
		}
		enc.Flush()
	default:
		return sendError(c, fmt.Errorf("unsupported format: %s", format))
	}
	return nil
}
