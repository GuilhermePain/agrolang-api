package worker

import (
	"context"
	"sync"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type PropertyLister interface {
	ListAll(ctx context.Context) ([]model.Property, error)
}

type ForecastFetcher interface {
	FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error)
}

type ScanResult struct {
	Property model.Property
	Result   risk.Result
	Err      error
}

func Scan(ctx context.Context, lister PropertyLister, fetcher ForecastFetcher, concurrency int) ([]ScanResult, error) {
	properties, err := lister.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]ScanResult, len(properties))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, p := range properties {
		wg.Add(1)
		go func(i int, p model.Property) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			readings, err := fetcher.FetchForecast(ctx, weather.Coordinates{Lat: p.Latitude, Lon: p.Longitude})
			if err != nil {
				results[i] = ScanResult{Property: p, Err: err}
				return
			}

			result := risk.Evaluate(risk.Property{CropStage: risk.CropStage(p.CropStage)}, readings)
			results[i] = ScanResult{Property: p, Result: result}
		}(i, p)
	}

	wg.Wait()
	return results, nil
}
