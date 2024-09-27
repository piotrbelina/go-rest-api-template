package server

import (
	"log/slog"
	"net/http"
)

type Config struct {
	Host string
	Port string
}

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	return mux
}
