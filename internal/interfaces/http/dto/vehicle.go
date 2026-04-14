package dto

type CreateVehicleRequest struct {
	Plate           string  `json:"plate" binding:"required"`
	Brand           string  `json:"brand" binding:"required"`
	Model           string  `json:"model" binding:"required"`
	ManufactureYear *int    `json:"manufacture_year"`
	ModelYear       int     `json:"model_year" binding:"required"`
	Color           *string `json:"color"`
	Mileage         *int    `json:"mileage"`
	Chassis         *string `json:"chassis"`
	Notes           *string `json:"notes"`
}

type VehicleResponse struct {
	ID              string  `json:"id"`
	ClientID        string  `json:"client_id"`
	Plate           string  `json:"plate"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear *int    `json:"manufacture_year,omitempty"`
	ModelYear       int     `json:"model_year"`
	Color           *string `json:"color,omitempty"`
	Mileage         *int    `json:"mileage,omitempty"`
	Chassis         *string `json:"chassis,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}
