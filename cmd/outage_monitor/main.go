package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/doesnotcommit/outage_monitor/internal/handlers"
	"github.com/doesnotcommit/outage_monitor/internal/outage"
	"github.com/doesnotcommit/outage_monitor/internal/parser"
	"github.com/doesnotcommit/outage_monitor/internal/plugin"
	"github.com/doesnotcommit/outage_monitor/internal/repo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	LogLevel              int `default:"-4"`
	DynamoAccessKey       string
	DynamoSecretAccessKey string
	DynamoRegion          string
}

func main() {
	var cfg config
	if err := aconfig.LoaderFor(&cfg, aconfig.Config{
		EnvPrefix: "OUTAGE_MONITOR",
	}).Load(); err != nil {
		slog.Error("load config", slog.Any("error", err))
		os.Exit(1)
	}
	sl := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.Level(cfg.LogLevel),
	}))
	ctx := context.Background()
	if err := run(ctx, cfg, sl); err != nil {
		sl.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config, sl *slog.Logger) error {
	handleErr := func(err error) error {
		return fmt.Errorf("run: %w", err)
	}
	ctx, cancelCtx := signal.NotifyContext(ctx, os.Interrupt)
	defer cancelCtx()
	handleWater, err := injectWater(ctx, cfg, sl)
	if err != nil {
		return handleErr(err)
	}
	handlers := map[string]http.HandlerFunc{
		"/water": handleWater,
	}
	handleHTTP(ctx, handlers, sl)
	return nil
}

func injectWater(ctx context.Context, cfg config, sl *slog.Logger) (http.HandlerFunc, error) {
	handleErr := func(err error) (http.HandlerFunc, error) {
		return nil, fmt.Errorf("inject water: %w", err)
	}
	ctx, cancelCtx := signal.NotifyContext(ctx, os.Interrupt)
	defer cancelCtx()
	waterGovGeParser, err := parser.NewWaterGovGe(sl)
	if err != nil {
		return handleErr(err)
	}
	waterGovGePlugin := plugin.NewWaterGovGe(waterGovGeParser)
	dynamo, err := repo.NewDynamoWaterGovGe(ctx, cfg.DynamoAccessKey, cfg.DynamoSecretAccessKey, cfg.DynamoRegion, time.Now, sl)
	if err != nil {
		return handleErr(err)
	}
	s := outage.NewService(ctx, waterGovGePlugin, dynamo, time.Hour, sl)
	go s.StartRefreshingData(ctx)
	h := handlers.NewHTTP(s, sl)
	handlers := map[string]http.HandlerFunc{
		"/water": h.HandleWater,
	}
	handleHTTP(ctx, handlers, sl)
	return h.HandleWater, nil
}

func handleHTTP(ctx context.Context, handlers map[string]http.HandlerFunc, sl *slog.Logger) {
	handleErr := func(err error) {
		sl.Error(err.Error())
	}
	mux := http.NewServeMux()
	for p, h := range handlers {
		mux.Handle(p, h)
	}
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go newSrvShutdown(ctx, &srv, sl)()
	if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		sl.Info("good bye")
	} else if err != nil {
		handleErr(err)
		return
	}
}

func handlePrometheus(ctx context.Context, sl *slog.Logger) {
	handleErr := func(err error) {
		sl.Error(err.Error())
	}
	reg := prometheus.NewPedanticRegistry()
	mux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:            newErrorLogger(sl),
		Registry:            reg,
		MaxRequestsInFlight: 32,
		Timeout:             time.Second,
	})
	mux.Handle("/metrics", promHandler)
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go newSrvShutdown(ctx, &srv, sl)()
	if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		sl.Info("good bye")
	} else if err != nil {
		handleErr(err)
		return
	}
}

type ErrorLogger func(v ...any)

func (el ErrorLogger) Println(v ...any) {
	el(v...)
}

func newErrorLogger(sl *slog.Logger) ErrorLogger {
	return func(v ...any) {
		sl.Error(fmt.Sprintln(v...))
	}
}

func newSrvShutdown(ctx context.Context, srv *http.Server, sl *slog.Logger) func() {
	return func() {
		<-ctx.Done()
		shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelShutdownCtx()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			sl.Error(err.Error())
		}
	}
}
