package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct {
	logger *slog.Logger
	tracer trace.Tracer
	meter  metric.Meter
}

func NewServer(logger *slog.Logger, tracer trace.Tracer, meter metric.Meter) *Server {
	return &Server{
		logger,
		tracer,
		meter,
	}
}

// (GET /ping)
func (s *Server) GetPing(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "GetPing")
	defer span.End()

	resp := Pong{
		Ping: "pong",
	}

	s.logger.InfoContext(ctx, "msg")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
