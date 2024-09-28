package server

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

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
	// 	tracer  = otel.Tracer(name)
	// 	meter   = otel.Meter(name)
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
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc("/", handleRolldice(logger, tracer, meter))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")

	return handler
}
