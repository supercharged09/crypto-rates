package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/repository"
)

// ChartService — сервис для генерации графиков
type ChartService struct {
	rateRepo repository.RateRepositoryInterface
}

// NewChartService создает новый сервис графиков
func NewChartService(rateRepo repository.RateRepositoryInterface) *ChartService {
	return &ChartService{rateRepo: rateRepo}
}

// GenerateChartHTML возвращает html строку с графиком
func (s *ChartService) GenerateChartHTML(ctx context.Context, cryptocurrency string) (string, error) {
	points, err := s.rateRepo.GetPrices24h(ctx, cryptocurrency)
	if err != nil {
		return "", fmt.Errorf("failed to get prices: %w", err)
	}

	if len(points) == 0 {
		return "", fmt.Errorf("no data for chart")
	}

	info := model.GetCryptoInfo(cryptocurrency)

	var xAxis []string
	var yAxis []opts.LineData

	for _, p := range points {
		xAxis = append(xAxis, p.Timestamp.Format("15:04"))
		yAxis = append(yAxis, opts.LineData{Value: p.Price})
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "800px",
			Height: "400px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s %s / USD — 24 часа", info.Emoji, info.Name),
			Subtitle: time.Now().Format("02.01.2006 15:04"),
		}),
	)

	line.SetXAxis(xAxis).AddSeries("Цена", yAxis)

	var buf bytes.Buffer
	line.Render(&buf)

	return buf.String(), nil
}
