package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
)

type Prediction struct {
	entity.Base
	WindowDays       int              `json:"window_days"`
	Recommendations  []Recommendation `json:"recommendations"`
	NotificationSent bool             `json:"notification_sent"`
}
