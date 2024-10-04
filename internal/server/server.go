package server

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/piotrbelina/go-rest-api-template/api"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Host string
	Port string
}

var (
	rollCnt metric.Int64Counter
)

func handleRolldice(logger *slog.Logger, tracer trace.Tracer, meter metric.Meter) func(w http.ResponseWriter, r *http.Request) {
	var err error
	rollCnt, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "roll")
		defer span.End()

		roll := 1 + rand.Intn(6)

		var msg string
		if player := r.PathValue("player"); player != "" {
			msg = fmt.Sprintf("%s is rolling the dice", player)
		} else {
			msg = "Anonymous player is rolling the dice"
		}
		logger.InfoContext(ctx, msg, "result", roll)

		rollValueAttr := attribute.Int("roll.value", roll)
		span.SetAttributes(rollValueAttr)
		rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

		resp := strconv.Itoa(roll) + "\n"
		if _, err := io.WriteString(w, resp); err != nil {
			logger.ErrorContext(ctx, "Write failed", "error", err)
		}
	}
}

func NewServer(logger *slog.Logger, tracer trace.Tracer, meter metric.Meter) http.Handler {
	mux := &MyMux{}

	server := api.NewServer(logger, tracer, meter)

	h := api.HandlerFromMux(server, mux)

	return h
}

type MyMux struct {
	http.ServeMux
}

func (r *MyMux) HandleFunc(pattern string, h func(http.ResponseWriter, *http.Request)) {
	handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(h))
	r.Handle(pattern, handler)
}
