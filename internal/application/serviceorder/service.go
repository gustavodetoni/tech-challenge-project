package serviceorder

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brazilian-utils/go/licenseplate"
	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/document"
	"github.com/soat-architecture/tech-challenge-project/pkg/br/plate"
)

var (
	ErrInvalidInput = expections.New(expections.CodeValidation, "invalid input")
)

type FlowUseCase struct {
	clients  repository.ClientRepository
	vehicles repository.VehicleRepository
	services repository.ServiceRepository
	parts    repository.PartRepository
	flow     repository.ServiceOrderFlowRepository
}

func NewFlowUseCase(
	clients repository.ClientRepository,
	vehicles repository.VehicleRepository,
	services repository.ServiceRepository,
	parts repository.PartRepository,
	flow repository.ServiceOrderFlowRepository,
) *FlowUseCase {
	return &FlowUseCase{
		clients:  clients,
		vehicles: vehicles,
		services: services,
		parts:    parts,
		flow:     flow,
	}
}

func NewService(
	clients repository.ClientRepository,
	vehicles repository.VehicleRepository,
	services repository.ServiceRepository,
	parts repository.PartRepository,
	flow repository.ServiceOrderFlowRepository,
) *FlowUseCase {
	return NewFlowUseCase(clients, vehicles, services, parts, flow)
}

type Service = FlowUseCase

type CreateDraftInput struct {
	ClientDocumentType   string
	ClientDocumentNumber string
	ClientEmail          *string
	ClientPhone          *string

	VehiclePlate           string
	VehicleManufactureYear *int
	VehicleColor           *string

	CustomerComplaint *string

	Services []ItemInput
	Parts    []ItemInput
}

type ItemInput struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type CreateDraftOutput struct {
	ServiceOrderID string
	Code           string
	BudgetID       string
	BudgetStatus   order.BudgetStatus
	TotalCents     int64
}

type ReviseBudgetInput struct {
	Services []ItemInput
	Parts    []ItemInput
}

type ReviseBudgetOutput struct {
	BudgetID     string
	BudgetStatus order.BudgetStatus
	Version      int
	TotalCents   int64
}

func (s *FlowUseCase) CreateDraft(ctx context.Context, in CreateDraftInput) (*CreateDraftOutput, error) {
	doc := strings.TrimSpace(in.ClientDocumentNumber)
	doc = document.Normalize(doc)
	if !isValidDocument(client.DocumentType(in.ClientDocumentType), doc) {
		return nil, fmt.Errorf("%w: invalid document", ErrInvalidInput)
	}

	vehiclePlate := plate.Normalize(in.VehiclePlate)
	if !licenseplate.IsValid(vehiclePlate, "") {
		return nil, fmt.Errorf("%w: invalid plate", ErrInvalidInput)
	}

	cl, err := s.clients.FindByDocument(ctx, doc)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: client not found", ErrInvalidInput)
		}
		return nil, err
	}

	if in.ClientEmail != nil || in.ClientPhone != nil {
		updated := false
		if in.ClientEmail != nil {
			cl.Email = in.ClientEmail
			updated = true
		}
		if in.ClientPhone != nil {
			cl.Phone = in.ClientPhone
			updated = true
		}
		if updated {
			cl.UpdatedAt = time.Now().UTC()
			if err := s.clients.Update(ctx, cl); err != nil {
				return nil, err
			}
		}
	}

	v, err := s.vehicles.FindByPlate(ctx, vehiclePlate)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: vehicle not found", ErrInvalidInput)
		}
		return nil, err
	}
	if v.ClientID != cl.ID {
		return nil, fmt.Errorf("%w: vehicle already belongs to another client", repository.ErrConflict)
	}

	if in.VehicleManufactureYear != nil || in.VehicleColor != nil {
		updated := false
		if in.VehicleManufactureYear != nil {
			v.ManufactureYear = in.VehicleManufactureYear
			updated = true
		}
		if in.VehicleColor != nil {
			v.Color = in.VehicleColor
			updated = true
		}
		if updated {
			v.UpdatedAt = time.Now().UTC()
			if err := s.vehicles.Update(ctx, v); err != nil {
				return nil, err
			}
		}
	}

	serviceLines, serviceTotal, err := s.buildServiceLines(ctx, in.Services)
	if err != nil {
		return nil, err
	}
	partLines, partTotal, err := s.buildPartLines(ctx, in.Parts)
	if err != nil {
		return nil, err
	}

	total := serviceTotal + partTotal
	now := time.Now().UTC()

	code, err := newServiceOrderCode()
	if err != nil {
		return nil, err
	}

	so := order.ServiceOrder{
		ID:                uuid.NewString(),
		Code:              code,
		ClientID:          cl.ID,
		VehicleID:         v.ID,
		Status:            order.StatusReceived,
		CustomerComplaint: in.CustomerComplaint,
		OpenedAt:          now,
	}
	b := order.Budget{
		ID:               uuid.NewString(),
		ServiceOrderID:   so.ID,
		Version:          1,
		Status:           order.BudgetStatusDraft,
		TotalAmountCents: total,
	}

	createdSO, createdBudget, err := s.flow.CreateDraft(ctx, repository.CreateServiceOrderDraftParams{
		ServiceOrder:   so,
		Budget:         b,
		BudgetServices: serviceLines,
		BudgetParts:    partLines,
	})
	if err != nil {
		return nil, err
	}

	return &CreateDraftOutput{
		ServiceOrderID: createdSO.ID,
		Code:           createdSO.Code,
		BudgetID:       createdBudget.ID,
		BudgetStatus:   createdBudget.Status,
		TotalCents:     createdBudget.TotalAmountCents,
	}, nil
}

func (s *FlowUseCase) ReviseBudget(ctx context.Context, serviceOrderID string, in ReviseBudgetInput, changedByUserID *string) (*ReviseBudgetOutput, error) {
	if serviceOrderID == "" {
		return nil, fmt.Errorf("%w: missing service_order_id", ErrInvalidInput)
	}

	serviceLines, serviceTotal, err := s.buildServiceLines(ctx, in.Services)
	if err != nil {
		return nil, err
	}
	partLines, partTotal, err := s.buildPartLines(ctx, in.Parts)
	if err != nil {
		return nil, err
	}
	if len(serviceLines) == 0 && len(partLines) == 0 {
		return nil, fmt.Errorf("%w: at least one item is required", ErrInvalidInput)
	}

	total := serviceTotal + partTotal

	now := time.Now().UTC()
	b := order.Budget{
		ID:               uuid.NewString(),
		ServiceOrderID:   serviceOrderID,
		Version:          0,
		Status:           order.BudgetStatusDraft,
		TotalAmountCents: total,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	createdBudget, err := s.flow.CreateBudgetRevision(ctx, repository.CreateBudgetRevisionParams{
		ServiceOrderID:  serviceOrderID,
		Budget:          b,
		BudgetServices:  serviceLines,
		BudgetParts:     partLines,
		ChangedByUserID: changedByUserID,
	})
	if err != nil {
		return nil, err
	}

	return &ReviseBudgetOutput{
		BudgetID:     createdBudget.ID,
		BudgetStatus: createdBudget.Status,
		Version:      createdBudget.Version,
		TotalCents:   createdBudget.TotalAmountCents,
	}, nil
}

func (s *FlowUseCase) StartDiagnosis(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return s.flow.StartDiagnosis(ctx, serviceOrderID, changedByUserID)
}

func (s *FlowUseCase) SendBudget(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return s.flow.SendLatestBudget(ctx, serviceOrderID, changedByUserID)
}

func (s *FlowUseCase) Finish(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return s.flow.Finish(ctx, serviceOrderID, changedByUserID)
}

func (s *FlowUseCase) Deliver(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return s.flow.Deliver(ctx, serviceOrderID, changedByUserID)
}

func (s *FlowUseCase) ClientGetByCode(ctx context.Context, code string, documentNumber string) (*repository.ClientServiceOrderView, error) {
	return s.flow.GetClientViewByCode(ctx, code, document.Normalize(documentNumber))
}

func (s *FlowUseCase) ClientApproveBudget(ctx context.Context, code string, documentNumber string) error {
	normalizedDoc := document.Normalize(documentNumber)
	existing, err := s.clients.FindByDocument(ctx, normalizedDoc)
	if err != nil {
		return err
	}

	var approvedByName *string
	if strings.TrimSpace(existing.Name) != "" {
		approvedByName = &existing.Name
	}
	return s.flow.ApproveLatestBudgetByCode(ctx, code, normalizedDoc, approvedByName)
}

func (s *FlowUseCase) ClientRejectBudget(ctx context.Context, code string, documentNumber string, reason string) error {
	return s.flow.RejectLatestBudgetByCode(ctx, code, document.Normalize(documentNumber), reason)
}

func (s *FlowUseCase) buildServiceLines(ctx context.Context, items []ItemInput) ([]order.BudgetServiceItem, int64, error) {
	if len(items) == 0 {
		return []order.BudgetServiceItem{}, 0, nil
	}

	ids := make([]string, 0, len(items))
	qtyByID := make(map[string]int, len(items))
	for _, it := range items {
		if it.ID == "" || it.Quantity <= 0 {
			return nil, 0, fmt.Errorf("%w: invalid service item", ErrInvalidInput)
		}
		ids = append(ids, it.ID)
		qtyByID[it.ID] += it.Quantity
	}

	svcs, err := s.services.FindByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	if len(svcs) != len(unique(ids)) {
		return nil, 0, fmt.Errorf("%w: service not found", repository.ErrNotFound)
	}

	lines := make([]order.BudgetServiceItem, 0, len(svcs))
	var total int64
	for _, svc := range svcs {
		q := qtyByID[svc.ID]
		lineTotal := svc.BasePriceCents * int64(q)
		lines = append(lines, order.BudgetServiceItem{
			ServiceID:       svc.ID,
			Description:     svc.Name,
			Quantity:        q,
			UnitPriceCents:  svc.BasePriceCents,
			TotalPriceCents: lineTotal,
		})
		total += lineTotal
	}
	return lines, total, nil
}

func (s *FlowUseCase) buildPartLines(ctx context.Context, items []ItemInput) ([]order.BudgetPartItem, int64, error) {
	if len(items) == 0 {
		return []order.BudgetPartItem{}, 0, nil
	}

	ids := make([]string, 0, len(items))
	qtyByID := make(map[string]int, len(items))
	for _, it := range items {
		if it.ID == "" || it.Quantity <= 0 {
			return nil, 0, fmt.Errorf("%w: invalid part item", ErrInvalidInput)
		}
		ids = append(ids, it.ID)
		qtyByID[it.ID] += it.Quantity
	}

	pts, err := s.parts.FindByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	if len(pts) != len(unique(ids)) {
		return nil, 0, fmt.Errorf("%w: part not found", repository.ErrNotFound)
	}

	lines := make([]order.BudgetPartItem, 0, len(pts))
	var total int64
	for _, p := range pts {
		q := qtyByID[p.ID]
		lineTotal := p.UnitPriceCents * int64(q)
		lines = append(lines, order.BudgetPartItem{
			PartID:          p.ID,
			Description:     p.Name,
			Quantity:        q,
			UnitPriceCents:  p.UnitPriceCents,
			TotalPriceCents: lineTotal,
		})
		total += lineTotal
	}
	return lines, total, nil
}

func unique(ids []string) []string {
	set := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := set[id]; ok {
			continue
		}
		set[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func isValidDocument(docType client.DocumentType, doc string) bool {
	switch docType {
	case client.DocumentTypeCPF:
		return document.IsValidCPF(doc)
	case client.DocumentTypeCNPJ:
		return document.IsValidCNPJ(doc)
	default:
		return false
	}
}

func newServiceOrderCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	suffix := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("OS-%s-%s", time.Now().UTC().Format("20060102"), suffix), nil
}

var _ = part.Part{}
var _ = service.Service{}
