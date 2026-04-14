package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/mapper"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminVehiclesController struct {
	svc *admin.VehicleService
}

func NewAdminVehiclesController(svc *admin.VehicleService) *AdminVehiclesController {
	return &AdminVehiclesController{svc: svc}
}

// @Summary List vehicles by client
// @Tags admin-vehicles
// @Param id path string true "Client ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.VehicleResponse
// @Router /admin/clients/{id}/vehicles [get]
func (h *AdminVehiclesController) ListByClient(c *gin.Context) {
	clientID := c.Param("id")
	limit, offset := shared.ParseLimitOffset(c)
	vehicles, err := h.svc.ListByClientID(c.Request.Context(), clientID, limit, offset)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}

	out := make([]dto.VehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		out = append(out, mapper.VehicleResponseFromDomain(v))
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Create vehicle for client
// @Tags admin-vehicles
// @Param id path string true "Client ID"
// @Param request body dto.CreateVehicleRequest true "Vehicle"
// @Success 201 {object} dto.VehicleResponse
// @Router /admin/clients/{id}/vehicles [post]
func (h *AdminVehiclesController) CreateForClient(c *gin.Context) {
	clientID := c.Param("id")
	var req dto.CreateVehicleRequest
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
	c.JSON(http.StatusCreated, mapper.VehicleResponseFromDomain(*v))
}

// @Summary Get vehicle
// @Tags admin-vehicles
// @Param id path string true "Vehicle ID"
// @Success 200 {object} dto.VehicleResponse
// @Router /admin/vehicles/{id} [get]
func (h *AdminVehiclesController) Get(c *gin.Context) {
	id := c.Param("id")
	v, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.VehicleResponseFromDomain(*v))
}

// @Summary Update vehicle
// @Tags admin-vehicles
// @Param id path string true "Vehicle ID"
// @Param request body dto.CreateVehicleRequest true "Vehicle"
// @Success 200 {object} dto.VehicleResponse
// @Router /admin/vehicles/{id} [put]
func (h *AdminVehiclesController) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	current, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
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
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.VehicleResponseFromDomain(*v))
}

// @Summary Delete vehicle
// @Tags admin-vehicles
// @Param id path string true "Vehicle ID"
// @Success 204
// @Router /admin/vehicles/{id} [delete]
func (h *AdminVehiclesController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
