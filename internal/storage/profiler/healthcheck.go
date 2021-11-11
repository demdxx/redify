package profiler

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/demdxx/redify/internal/context/ctxlogger"
)

// HealthCheckHandler of service
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status":"OK"}`)); err != nil {
		ctxlogger.Get(r.Context()).Error("write HTTP response", zap.Error(err))
	}
}
