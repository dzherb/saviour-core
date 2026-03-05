package model

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UUID          uuid.UUID
	CreatedAt     time.Time
	UserUUID      uuid.UUID
	LastRefreshAt time.Time
	UserAgent     string
	IP            netip.Addr
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
