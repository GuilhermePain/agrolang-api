package worker_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
	"github.com/GuilhermePain/agrolang-api/internal/worker"
)

type fakePropertyLister struct {
	properties []model.Property
}

func (f *fakePropertyLister) ListAll(ctx context.Context) ([]model.Property, error) {
	return f.properties, nil
}

type fakeForecastFetcher struct {
	mu            sync.Mutex
	calls         atomic.Int32
	maxConcurrent int32
	current       atomic.Int32
	readings      []risk.WeatherReading
	err           error
}

func (f *fakeForecastFetcher) FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error) {
	f.calls.Add(1)
	cur := f.current.Add(1)
	defer f.current.Add(-1)

	f.mu.Lock()
	if cur > f.maxConcurrent {
		f.maxConcurrent = cur
	}
	f.mu.Unlock()

	time.Sleep(5 * time.Millisecond)

	if f.err != nil {
		return nil, f.err
	}
	return f.readings, nil
}

func newProperties(n int) []model.Property {
	props := make([]model.Property, n)
	for i := range props {
		props[i] = model.Property{
			ID:        "prop-" + string(rune('a'+i)),
			CropStage: model.CropStageFlowering,
			Latitude:  -22.9,
			Longitude: -47.06,
		}
	}
	return props
}

func TestScan_EvaluatesRiskForEveryProperty(t *testing.T) {
	lister := &fakePropertyLister{properties: newProperties(3)}
	fetcher := &fakeForecastFetcher{
		readings: []risk.WeatherReading{
			{Time: time.Now(), TempAvgC: 20, HumidityPct: 60},
		},
	}

	results, err := worker.Scan(context.Background(), lister, fetcher, 2)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected per-property error: %v", r.Err)
		}
		if r.Result.Level != risk.LevelLow {
			t.Errorf("expected LevelLow, got %v", r.Result.Level)
		}
	}
}

func TestScan_RespectsConcurrencyLimit(t *testing.T) {
	lister := &fakePropertyLister{properties: newProperties(10)}
	fetcher := &fakeForecastFetcher{
		readings: []risk.WeatherReading{{Time: time.Now(), TempAvgC: 20, HumidityPct: 60}},
	}

	const limit = 3
	_, err := worker.Scan(context.Background(), lister, fetcher, limit)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if fetcher.calls.Load() != 10 {
		t.Errorf("expected 10 fetch calls, got %d", fetcher.calls.Load())
	}
	if fetcher.maxConcurrent > limit {
		t.Errorf("expected max concurrency <= %d, got %d", limit, fetcher.maxConcurrent)
	}
}

func TestScan_PerPropertyFetchError_DoesNotAbortOtherProperties(t *testing.T) {
	lister := &fakePropertyLister{properties: newProperties(2)}
	fetcher := &fakeForecastFetcher{err: errors.New("api unavailable")}

	results, err := worker.Scan(context.Background(), lister, fetcher, 2)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err == nil {
			t.Error("expected per-property error, got nil")
		}
	}
}
