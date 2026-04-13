package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
)

type AdminVehiclesController struct {
	svc *admin.VehicleService
}

func NewAdminVehiclesController(svc *admin.VehicleService) *AdminVehiclesController {
	return &AdminVehiclesController{svc: svc}
}

type createVehicleRequest struct {
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

type vehicleResponse struct {
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

// AdminListVehiclesByClient godoc
// @Summary List vehicles by client
// @Tags admin-vehicles
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} vehicleResponse
// @Router /admin/clients/{id}/vehicles [get]
func (h *AdminVehiclesController) ListByClient(c *gin.Context) {
	clientID := c.Param("id")
	limit, offset := parseLimitOffset(c)
	vehicles, err := h.svc.ListByClientID(c.Request.Context(), clientID, limit, offset)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	out := make([]vehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		out = append(out, mapVehicleResponse(v))
	}
	c.JSON(http.StatusOK, out)
}

// AdminCreateVehicle godoc
// @Summary Create vehicle for client
// @Tags admin-vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body createVehicleRequest true "Vehicle"
// @Success 201 {object} vehicleResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /admin/clients/{id}/vehicles [post]
func (h *AdminVehiclesController) CreateForClient(c *gin.Context) {
	clientID := c.Param("id")
	var req createVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	v, err := h.svc.Create(c.Request.Context(), vehicle.Vehicle{
		ClientID:        clientID,
		Plate:           req.Plate,
		Brand:           req.Brand,
		Model:           req.Model,
		ManufactureYear: req.ManufactureYear,
		ModelYear:       req.ModelYear,
		Color:           req.Color,
		Mileage:         req.Mileage,
		Chassis:         req.Chassis,
		Notes:           req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mapVehicleResponse(*v))
}

// AdminGetVehicle godoc
// @Summary Get vehicle
// @Tags admin-vehicles
// @Security BearerAuth
// @Param id path string true "Vehicle ID"
// @Success 200 {object} vehicleResponse
// @Failure 404 {object} map[string]string
// @Router /admin/vehicles/{id} [get]
func (h *AdminVehiclesController) Get(c *gin.Context) {
	id := c.Param("id")
	v, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapVehicleResponse(*v))
}

// AdminUpdateVehicle godoc
// @Summary Update vehicle
// @Tags admin-vehicles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param request body createVehicleRequest true "Vehicle"
// @Success 200 {object} vehicleResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/vehicles/{id} [put]
func (h *AdminVehiclesController) Update(c *gin.Context) {
	id := c.Param("id")
	var req createVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	current, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	v, err := h.svc.Update(c.Request.Context(), id, vehicle.Vehicle{
		ClientID:        current.ClientID,
		Plate:           req.Plate,
		Brand:           req.Brand,
		Model:           req.Model,
		ManufactureYear: req.ManufactureYear,
		ModelYear:       req.ModelYear,
		Color:           req.Color,
		Mileage:         req.Mileage,
		Chassis:         req.Chassis,
		Notes:           req.Notes,
	})
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapVehicleResponse(*v))
}

// AdminDeleteVehicle godoc
// @Summary Delete vehicle
// @Tags admin-vehicles
// @Security BearerAuth
// @Param id path string true "Vehicle ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /admin/vehicles/{id} [delete]
func (h *AdminVehiclesController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mapVehicleResponse(v vehicle.Vehicle) vehicleResponse {
	return vehicleResponse{
		ID:              v.ID,
		ClientID:        v.ClientID,
		Plate:           v.Plate,
		Brand:           v.Brand,
		Model:           v.Model,
		ManufactureYear: v.ManufactureYear,
		ModelYear:       v.ModelYear,
		Color:           v.Color,
		Mileage:         v.Mileage,
		Chassis:         v.Chassis,
		Notes:           v.Notes,
	}
}
