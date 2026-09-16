package model

import "time"

type CropStage string

const (
	CropStageGermination CropStage = "germination"
	CropStageFlowering   CropStage = "flowering"
	CropStageHarvest     CropStage = "harvest"
)

type Property struct {
	ID         string
	ProducerID string
	Latitude   float64
	Longitude  float64
	Crop       string
	SoilType   string
	CropStage  CropStage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
