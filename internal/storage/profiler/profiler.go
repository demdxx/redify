package profiler

import (
	"fmt"
	"net/http"

	"github.com/pkg/profile"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Run profiler
func Run(mode, listenAddr string, logger *zap.Logger, withDefHandlers ...bool) {
	switch mode {
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "mem", "memory":
		defer profile.Start(profile.MemProfile).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile).Stop()
	case "net":
		go func() {
			fmt.Printf("Run profile (port %s)\n", listenAddr)
			if len(withDefHandlers) > 0 && withDefHandlers[0] {
				http.HandleFunc("/healthcheck", HealthCheckHandler)
				http.Handle("/metrics", promhttp.Handler())
			}
			if err := http.ListenAndServe(listenAddr, nil); err != nil {
				logger.Error("profile server error", zap.Error(err))
			}
		}()
	}
}
