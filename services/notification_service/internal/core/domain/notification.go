package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
)

type Notification struct {
	entity.Base
	EventID string `json:"event_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Read    bool   `json:"read"`
}
