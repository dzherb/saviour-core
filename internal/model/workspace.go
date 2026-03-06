package model

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	UUID       uuid.UUID
	CreatedAt  time.Time
	Name       string
	AuthorUUID uuid.UUID
}
