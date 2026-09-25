package model

import "time"

type CropStage string

const (
	CropStageGermination CropStage = "germination"
	CropStageFlowering   CropStage = "flowering"
	CropStageHarvest     CropStage = "harvest"
)

type Property struct {
	ID         string    `json:"id"`
	ProducerID string    `json:"producer_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Crop       string    `json:"crop"`
	SoilType   string    `json:"soil_type"`
	CropStage  CropStage `json:"crop_stage"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
