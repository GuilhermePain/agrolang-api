package worker_test

import (
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
	"github.com/GuilhermePain/agrolang-api/internal/worker"
)

var (
	_ worker.ForecastFetcher = (*weather.OpenMeteoClient)(nil)
	_ worker.PropertyLister  = (*repository.PropertyRepository)(nil)
)
