package handlers

import (
	"context"
	"log/slog"
	"net/http"
)

type OutageMonitor interface {
	GetWaterOutages(ctx context.Context) error
}

type HTTP struct {
	omon OutageMonitor
	sl   *slog.Logger
}

func NewHTTP(omon OutageMonitor, sl *slog.Logger) HTTP {
	return HTTP{omon, sl}
}

func (h HTTP) HandleWater(res http.ResponseWriter, req *http.Request) {
	h.sl.Info("handling a water request")
	res.WriteHeader(http.StatusOK)
}
