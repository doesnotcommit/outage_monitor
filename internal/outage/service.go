package outage

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type WaterGovGePlugin interface {
	GetWaterOutages(ctx context.Context) ([]WaterGovGe, error)
}

type WaterGovGeRepo interface {
	SaveOutages(ctx context.Context, outages ...WaterGovGe) error
}

type Service struct {
	plugin WaterGovGePlugin
	repo   WaterGovGeRepo
	ticker *time.Ticker
	sl     *slog.Logger
}

func NewService(ctx context.Context, parser WaterGovGePlugin, repo WaterGovGeRepo, interval time.Duration, sl *slog.Logger) Service {
	ticker := time.NewTicker(interval)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	return Service{parser, repo, ticker, sl}
}

func (s Service) StartRefreshingData(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.sl.Info("stopping periodic refresh", slog.Any("ctx error", ctx.Err()))
			return
		default:
			if err := s.refreshWaterOutages(ctx); err != nil {
				s.sl.Error("refresh water outages", slog.Any("err", err))
			}
			<-s.ticker.C
		}
	}
}

func (s Service) refreshWaterOutages(ctx context.Context) error {
	handleErr := func(err error) error {
		return fmt.Errorf("refresh water outages: %w", err)
	}
	waterOutages, err := s.plugin.GetWaterOutages(ctx)
	if err != nil {
		return handleErr(err)
	}
	if err := s.repo.SaveOutages(ctx, waterOutages...); err != nil {
		return handleErr(err)
	}
	return nil
}

func (s Service) GetWaterOutages(ctx context.Context) error {
	return nil
}
