package model

import (
	"time"

	"github.com/google/uuid"
)

type Secret struct {
	UUID       uuid.UUID
	CreatedAt  time.Time
	Name       string
	AuthorUUID uuid.UUID
}
