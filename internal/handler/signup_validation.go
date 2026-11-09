package handler

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

var (
	phonePattern = regexp.MustCompile(`^55\d{10,11}$`)
	statePattern = regexp.MustCompile(`^[A-Z]{2}$`)

	validConcerns = map[string]bool{
		"frost": true, "heat_wave": true, "heavy_rain": true, "severe_water_stress": true,
	}
	validSoils  = map[string]bool{"argiloso": true, "arenoso": true, "misto": true, "nao_informado": true}
	validStages = map[model.CropStage]bool{
		model.CropStageGermination: true, model.CropStageFlowering: true, model.CropStageHarvest: true,
	}
	onboardingValues = map[string]map[string]bool{
		"work_style":       {"family": true, "multi_area": true, "technician": true},
		"areas_bucket":     {"1": true, "2-5": true, "6+": true},
		"producers_bucket": {"ate-10": true, "11-40": true, "mais-40": true},
	}
)

type signupRequest struct {
	Profile       model.Profile     `json:"profile"`
	Name          string            `json:"name"`
	WhatsAppPhone string            `json:"whatsapp_phone"`
	Email         string            `json:"email"`
	Password      string            `json:"password"`
	Organization  *string           `json:"organization"`
	Concerns      []string          `json:"concerns"`
	Onboarding    map[string]string `json:"onboarding"`
	Producer      *struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"producer"`
	Property *struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		SoilType  string   `json:"soil_type"`
		// Crops share the location and soil; each one becomes its own property
		// so the risk engine applies that crop's thresholds and stage.
		Crops []struct {
			Crop      string          `json:"crop"`
			CropStage model.CropStage `json:"crop_stage"`
		} `json:"crops"`
	} `json:"property"`
}

const maxSignupCrops = 10

// validate returns user-facing (pt-BR) messages keyed by field; "producer.*"
// and "property.*" keys point at fields nested in those objects.
func (req *signupRequest) validate() map[string]string {
	fields := map[string]string{}

	if req.Profile != model.ProfileProducer && req.Profile != model.ProfileTechnician {
		fields["profile"] = "Escolha como você trabalha."
	}
	if len([]rune(strings.TrimSpace(req.Name))) < 2 {
		fields["name"] = "Digite seu nome."
	}
	if !phonePattern.MatchString(req.WhatsAppPhone) {
		fields["whatsapp_phone"] = "Digite o DDD e o número do WhatsApp."
	}
	if !validEmail(req.Email) {
		fields["email"] = "Digite um e-mail válido."
	}
	if n := len(req.Password); n < auth.MinPasswordLength || n > auth.MaxPasswordLength {
		fields["password"] = "A senha precisa ter de 8 a 72 caracteres."
	}
	for _, c := range req.Concerns {
		if !validConcerns[c] {
			fields["concerns"] = "Escolha riscos da lista."
			break
		}
	}
	for key, value := range req.Onboarding {
		allowed, known := onboardingValues[key]
		if !known || !allowed[value] {
			fields["onboarding"] = "As respostas do quiz estão inválidas."
			break
		}
	}

	organization := ""
	if req.Organization != nil {
		organization = strings.TrimSpace(*req.Organization)
	}

	switch req.Profile {
	case model.ProfileTechnician:
		if organization == "" {
			fields["organization"] = "Digite o nome da sua organização."
		}
		if req.Producer != nil {
			fields["producer"] = "Técnicos cadastram produtores depois, no painel."
		}
		if req.Property != nil {
			fields["property"] = "Técnicos cadastram propriedades depois, no painel."
		}
	case model.ProfileProducer:
		if organization != "" {
			fields["organization"] = "Organização é só para técnicos."
		}
		req.validateProducerAndProperty(fields)
	}

	return fields
}

func (req *signupRequest) validateProducerAndProperty(fields map[string]string) {
	if req.Producer == nil {
		fields["producer"] = "Informe a cidade."
	} else {
		if strings.TrimSpace(req.Producer.City) == "" {
			fields["producer.city"] = "Escolha a cidade."
		}
		if !statePattern.MatchString(strings.ToUpper(strings.TrimSpace(req.Producer.State))) {
			fields["producer.state"] = "Informe o estado com 2 letras."
		}
	}

	if req.Property == nil {
		fields["property"] = "Marque a sua terra no mapa."
		return
	}
	p := req.Property
	if p.Latitude == nil || *p.Latitude < -90 || *p.Latitude > 90 {
		fields["property.latitude"] = "Marque a sua terra no mapa."
	}
	if p.Longitude == nil || *p.Longitude < -180 || *p.Longitude > 180 {
		fields["property.longitude"] = "Marque a sua terra no mapa."
	}
	if !validSoils[p.SoilType] {
		fields["property.soil_type"] = "Escolha o tipo de solo."
	}
	if msg := cropsError(req); msg != "" {
		fields["property.crops"] = msg
	}
}

// cropsError returns the first problem with the crop list, or "" when it is valid.
func cropsError(req *signupRequest) string {
	crops := req.Property.Crops
	if len(crops) == 0 {
		return "Escolha pelo menos uma cultura."
	}
	if len(crops) > maxSignupCrops {
		return fmt.Sprintf("Escolha até %d culturas.", maxSignupCrops)
	}
	seen := map[string]bool{}
	for _, c := range crops {
		name := strings.ToLower(strings.TrimSpace(c.Crop))
		if name == "" {
			return "Escolha ou digite a cultura."
		}
		if !validStages[c.CropStage] {
			return "Escolha em que fase está cada cultura."
		}
		if seen[name] {
			return "Essa cultura está repetida."
		}
		seen[name] = true
	}
	return ""
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
