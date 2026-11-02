package server

import (
	"log/slog"
	"net/http"

	"github.com/forgeflow/usermanagement/internal/config"
	"github.com/forgeflow/usermanagement/internal/handler"
	"github.com/forgeflow/usermanagement/internal/middleware"
	"github.com/forgeflow/usermanagement/internal/service"
)

// NewRouter wires routes and middleware.
func NewRouter(svc *service.UserService, cfg config.Config, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	handler.NewUserHandler(svc).Register(mux)

	var h http.Handler = mux
	h = middleware.MaxBody(cfg.MaxBodySize)(h)
	h = middleware.CORS(cfg.CORSOrigin)(h)
	h = middleware.Logging(log)(h)
	h = middleware.Recover(log)(h)
	return h
}
