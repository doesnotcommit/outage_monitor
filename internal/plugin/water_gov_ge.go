package plugin

import (
	"context"

	"github.com/doesnotcommit/outage_monitor/internal/outage"
)

type WaterGovGeParser interface {
	GetOutages(ctx context.Context) ([]outage.WaterGovGe, error)
}

type WaterGovGe struct {
	waterGovGeParser WaterGovGeParser
}

func NewWaterGovGe(parser WaterGovGeParser) WaterGovGe {
	return WaterGovGe{parser}
}

func (w WaterGovGe) GetWaterOutages(ctx context.Context) ([]outage.WaterGovGe, error) {
	// TODO
	return w.waterGovGeParser.GetOutages(ctx)
}
