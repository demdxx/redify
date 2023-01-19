package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/demdxx/goconfig"
	"go.uber.org/zap"

	"github.com/demdxx/redify/cmd/appcontext"
	"github.com/demdxx/redify/cmd/server"
	"github.com/demdxx/redify/internal/context/ctxlogger"
	"github.com/demdxx/redify/internal/storage"
	"github.com/demdxx/redify/internal/storage/connect"
	"github.com/demdxx/redify/internal/storage/multistore"
	"github.com/demdxx/redify/internal/storage/profiler"
	"github.com/demdxx/redify/internal/storage/proxy"
	"github.com/demdxx/redify/internal/storage/simplecache"
	"github.com/demdxx/redify/internal/zlogger"
)

var (
	appVersion   string
	buildCommit  string
	buildVersion string
	buildDate    string
	config       appcontext.ConfigType
)

func init() {
	fatalError(goconfig.Load(&config), "load config:")

	if config.IsDebug() {
		fmt.Println(&config)
	}

	// Init new logger object
	loggerObj, err := zlogger.New(config.ServiceName, config.LogEncoder,
		config.LogLevel, config.LogAddr, zap.Fields(
			zap.String("commit", buildCommit),
			zap.String("version", appVersion),
			zap.String("build_version", buildVersion),
			zap.String("build_date", buildDate),
		))

	fatalError(err, "init logger")

	// Register global logger
	zap.ReplaceGlobals(loggerObj)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ctxlogger.WithLogger(ctx, zap.L())
	defer cancel()

	var (
		cache  storage.Cacher
		stores []storage.Driver
		err    error
	)

	// Connect global cache
	if config.Cache.Connect != "" {
		cache, err = simplecache.New(config.Cache.Size, config.Cache.TTL)
		fatalError(err, "create simple cache")
	}

	// Connect sources and bind redify keys
	for _, sconf := range config.Sources {
		st, err := connect.Connect(ctx, sconf.Connect)
		fatalError(err, sconf.Connect)
		if cacheSupport, ok := st.(storage.CacheSupporter); ok && !cacheSupport.SupportCache() {
			stores = append(stores, st)
		} else {
			stores = append(stores, proxy.New(ctx, cache, st, sconf.NotifyChannel))
		}
		for _, bind := range sconf.Binds {
			err = st.Bind(ctx, &storage.BindConfig{
				Pattern:     bind.Key,
				DBNum:       bind.DBNum,
				TableName:   bind.TableName,
				Readonly:    bind.Readonly,
				WhereExt:    bind.WhereExt,
				GetQuery:    bind.GetQuery,
				ListQuery:   bind.ListQuery,
				UpsertQuery: bind.UpsertQuery,
				DelQuery:    bind.DelQuery,
			})
			fatalError(err, sconf.Connect+" @ bind error")
		}
	}

	var (
		wg    sync.WaitGroup
		store = multistore.New(stores...)
	)

	defer func() {
		fatalError(store.Close(), "close DB connection")
	}()

	if config.Server.Profile.Listen != "" {
		profiler.Run(
			config.Server.Profile.Mode,
			config.Server.Profile.Listen,
			zap.L(),
			true)
	}

	if config.Server.RedisServer.Listen != "" {
		zap.L().Info("Run Redis server", zap.String("listen", config.Server.RedisServer.Listen))
		wg.Add(1)
		go func() {
			srv := server.RedisServer{
				RequestTimeout: config.Server.RedisServer.ReadTimeout,
				Driver:         store,
			}
			err := srv.ListenAndServe(ctx, config.Server.RedisServer.Listen)
			fatalError(err, "Listen Redis server")
			wg.Done()
		}()
	}

	if config.Server.HTTPServer.Listen != "" {
		zap.L().Info("Run HTTP server", zap.String("listen", config.Server.HTTPServer.Listen))
		wg.Add(1)
		go func() {
			srv := server.HTTPServer{
				RequestTimeout: config.Server.HTTPServer.ReadTimeout,
				Driver:         store,
			}
			err := srv.ListenAndServe(ctx, config.Server.HTTPServer.Listen)
			fatalError(err, "Listen HTTP server")
			wg.Done()
		}()
	}

	wg.Wait()
}

func fatalError(err error, msgs ...any) {
	if err != nil {
		log.Fatalln(append(msgs, err)...)
	}
}
