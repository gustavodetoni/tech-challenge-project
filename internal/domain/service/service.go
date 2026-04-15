package service

import "time"

type Service struct {
	ID               string
	Name             string
	Description      *string
	BasePriceCents   int64
	EstimatedMinutes int
	Active           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
