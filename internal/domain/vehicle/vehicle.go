package vehicle

import "time"

type Vehicle struct {
	ID              string
	ClientID        string
	Plate           string
	Brand           string
	Model           string
	ManufactureYear *int
	ModelYear       int
	Color           *string
	Mileage         *int
	Chassis         *string
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
