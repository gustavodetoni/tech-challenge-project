package repositories

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
)

type Store struct {
	mu sync.Mutex

	usersByID    map[string]*user.User
	usersByEmail map[string]*user.User

	clientsByID       map[string]*client.Client
	clientsByDocument map[string]*client.Client

	vehiclesByID    map[string]*vehicle.Vehicle
	vehiclesByPlate map[string]*vehicle.Vehicle

	servicesByID   map[string]*service.Service
	servicesByName map[string]*service.Service

	partsByID  map[string]*part.Part
	partsBySKU map[string]*part.Part

	serviceOrdersByID   map[string]*order.ServiceOrder
	serviceOrdersByCode map[string]*order.ServiceOrder

	budgetsByID            map[string]*order.Budget
	budgetsByServiceOrder  map[string][]*order.Budget
	budgetServicesByBudget map[string][]order.BudgetServiceItem
	budgetPartsByBudget    map[string][]order.BudgetPartItem

	statusHistoryBySO map[string][]order.StatusHistoryEntry
}

func NewStore() *Store {
	return &Store{
		usersByID:    map[string]*user.User{},
		usersByEmail: map[string]*user.User{},

		clientsByID:       map[string]*client.Client{},
		clientsByDocument: map[string]*client.Client{},

		vehiclesByID:    map[string]*vehicle.Vehicle{},
		vehiclesByPlate: map[string]*vehicle.Vehicle{},

		servicesByID:   map[string]*service.Service{},
		servicesByName: map[string]*service.Service{},

		partsByID:  map[string]*part.Part{},
		partsBySKU: map[string]*part.Part{},

		serviceOrdersByID:   map[string]*order.ServiceOrder{},
		serviceOrdersByCode: map[string]*order.ServiceOrder{},

		budgetsByID:            map[string]*order.Budget{},
		budgetsByServiceOrder:  map[string][]*order.Budget{},
		budgetServicesByBudget: map[string][]order.BudgetServiceItem{},
		budgetPartsByBudget:    map[string][]order.BudgetPartItem{},

		statusHistoryBySO: map[string][]order.StatusHistoryEntry{},
	}
}

// -------------------- Users --------------------

type UserRepository struct{ s *Store }

func NewUserRepository(s *Store) *UserRepository { return &UserRepository{s: s} }

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	email := strings.ToLower(strings.TrimSpace(u.Email))
	if email == "" {
		return errors.New("email is required")
	}
	if _, ok := r.s.usersByEmail[email]; ok {
		return repository.ErrConflict
	}

	cp := *u
	r.s.usersByID[cp.ID] = &cp
	r.s.usersByEmail[email] = &cp
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	email = strings.ToLower(strings.TrimSpace(email))
	u := r.s.usersByEmail[email]
	if u == nil {
		return nil, repository.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	u := r.s.usersByID[id]
	if u == nil {
		return nil, repository.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id string, role user.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	u := r.s.usersByID[id]
	if u == nil {
		return repository.ErrNotFound
	}
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// -------------------- Clients --------------------

type ClientRepository struct{ s *Store }

func NewClientRepository(s *Store) *ClientRepository { return &ClientRepository{s: s} }

func (r *ClientRepository) Create(ctx context.Context, c *client.Client) error {
	if c == nil {
		return errors.New("client is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	doc := strings.TrimSpace(c.DocumentNumber)
	if doc == "" {
		return errors.New("document_number is required")
	}
	if _, ok := r.s.clientsByDocument[doc]; ok {
		return repository.ErrConflict
	}
	cp := *c
	r.s.clientsByID[cp.ID] = &cp
	r.s.clientsByDocument[doc] = &cp
	return nil
}

func (r *ClientRepository) Update(ctx context.Context, c *client.Client) error {
	if c == nil {
		return errors.New("client is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.clientsByID[c.ID]
	if existing == nil {
		return repository.ErrNotFound
	}
	doc := strings.TrimSpace(c.DocumentNumber)
	if doc == "" {
		return errors.New("document_number is required")
	}
	if other := r.s.clientsByDocument[doc]; other != nil && other.ID != c.ID {
		return repository.ErrConflict
	}
	cp := *c
	r.s.clientsByID[cp.ID] = &cp
	r.s.clientsByDocument[doc] = &cp
	return nil
}

func (r *ClientRepository) Delete(ctx context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.clientsByID[id]
	if existing == nil {
		return repository.ErrNotFound
	}
	delete(r.s.clientsByDocument, existing.DocumentNumber)
	delete(r.s.clientsByID, id)
	return nil
}

func (r *ClientRepository) FindByID(ctx context.Context, id string) (*client.Client, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	c := r.s.clientsByID[id]
	if c == nil {
		return nil, repository.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *ClientRepository) FindByDocument(ctx context.Context, documentNumber string) (*client.Client, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	c := r.s.clientsByDocument[documentNumber]
	if c == nil {
		return nil, repository.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *ClientRepository) List(ctx context.Context, limit, offset int) ([]client.Client, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	items := make([]*client.Client, 0, len(r.s.clientsByID))
	for _, v := range r.s.clientsByID {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginate(items, limit, offset, func(c *client.Client) client.Client { return *c }), nil
}

// -------------------- Vehicles --------------------

type VehicleRepository struct{ s *Store }

func NewVehicleRepository(s *Store) *VehicleRepository { return &VehicleRepository{s: s} }

func (r *VehicleRepository) Create(ctx context.Context, v *vehicle.Vehicle) error {
	if v == nil {
		return errors.New("vehicle is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	plate := strings.TrimSpace(v.Plate)
	if plate == "" {
		return errors.New("plate is required")
	}
	if _, ok := r.s.vehiclesByPlate[plate]; ok {
		return repository.ErrConflict
	}
	cp := *v
	r.s.vehiclesByID[cp.ID] = &cp
	r.s.vehiclesByPlate[plate] = &cp
	return nil
}

func (r *VehicleRepository) Update(ctx context.Context, v *vehicle.Vehicle) error {
	if v == nil {
		return errors.New("vehicle is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.vehiclesByID[v.ID]
	if existing == nil {
		return repository.ErrNotFound
	}
	plate := strings.TrimSpace(v.Plate)
	if other := r.s.vehiclesByPlate[plate]; other != nil && other.ID != v.ID {
		return repository.ErrConflict
	}
	cp := *v
	r.s.vehiclesByID[cp.ID] = &cp
	r.s.vehiclesByPlate[plate] = &cp
	return nil
}

func (r *VehicleRepository) Delete(ctx context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.vehiclesByID[id]
	if existing == nil {
		return repository.ErrNotFound
	}
	delete(r.s.vehiclesByPlate, existing.Plate)
	delete(r.s.vehiclesByID, id)
	return nil
}

func (r *VehicleRepository) FindByID(ctx context.Context, id string) (*vehicle.Vehicle, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	v := r.s.vehiclesByID[id]
	if v == nil {
		return nil, repository.ErrNotFound
	}
	cp := *v
	return &cp, nil
}

func (r *VehicleRepository) FindByPlate(ctx context.Context, plate string) (*vehicle.Vehicle, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	v := r.s.vehiclesByPlate[plate]
	if v == nil {
		return nil, repository.ErrNotFound
	}
	cp := *v
	return &cp, nil
}

func (r *VehicleRepository) ListByClientID(ctx context.Context, clientID string, limit, offset int) ([]vehicle.Vehicle, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	items := make([]*vehicle.Vehicle, 0)
	for _, v := range r.s.vehiclesByID {
		if v.ClientID == clientID {
			items = append(items, v)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginate(items, limit, offset, func(v *vehicle.Vehicle) vehicle.Vehicle { return *v }), nil
}

// -------------------- Services --------------------

type ServiceRepository struct{ s *Store }

func NewServiceRepository(s *Store) *ServiceRepository { return &ServiceRepository{s: s} }

func (r *ServiceRepository) Create(ctx context.Context, s0 *service.Service) error {
	if s0 == nil {
		return errors.New("service is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	name := strings.TrimSpace(strings.ToLower(s0.Name))
	if name == "" {
		return errors.New("name is required")
	}
	if _, ok := r.s.servicesByName[name]; ok {
		return repository.ErrConflict
	}
	cp := *s0
	r.s.servicesByID[cp.ID] = &cp
	r.s.servicesByName[name] = &cp
	return nil
}

func (r *ServiceRepository) Update(ctx context.Context, s0 *service.Service) error {
	if s0 == nil {
		return errors.New("service is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.servicesByID[s0.ID]
	if existing == nil {
		return repository.ErrNotFound
	}
	name := strings.TrimSpace(strings.ToLower(s0.Name))
	if other := r.s.servicesByName[name]; other != nil && other.ID != s0.ID {
		return repository.ErrConflict
	}
	cp := *s0
	r.s.servicesByID[cp.ID] = &cp
	r.s.servicesByName[name] = &cp
	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.servicesByID[id]
	if existing == nil {
		return repository.ErrNotFound
	}
	delete(r.s.servicesByName, strings.ToLower(existing.Name))
	delete(r.s.servicesByID, id)
	return nil
}

func (r *ServiceRepository) FindByID(ctx context.Context, id string) (*service.Service, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	s0 := r.s.servicesByID[id]
	if s0 == nil {
		return nil, repository.ErrNotFound
	}
	cp := *s0
	return &cp, nil
}

func (r *ServiceRepository) FindByIDs(ctx context.Context, ids []string) ([]service.Service, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	out := make([]service.Service, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		s0 := r.s.servicesByID[id]
		if s0 == nil {
			continue
		}
		out = append(out, *s0)
	}
	return out, nil
}

func (r *ServiceRepository) List(ctx context.Context, limit, offset int) ([]service.Service, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	items := make([]*service.Service, 0, len(r.s.servicesByID))
	for _, v := range r.s.servicesByID {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginate(items, limit, offset, func(s0 *service.Service) service.Service { return *s0 }), nil
}

// -------------------- Parts --------------------

type PartRepository struct{ s *Store }

func NewPartRepository(s *Store) *PartRepository { return &PartRepository{s: s} }

func (r *PartRepository) Create(ctx context.Context, p *part.Part) error {
	if p == nil {
		return errors.New("part is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	sku := strings.TrimSpace(strings.ToLower(p.SKU))
	if sku == "" {
		return errors.New("sku is required")
	}
	if _, ok := r.s.partsBySKU[sku]; ok {
		return repository.ErrConflict
	}
	cp := *p
	r.s.partsByID[cp.ID] = &cp
	r.s.partsBySKU[sku] = &cp
	return nil
}

func (r *PartRepository) Update(ctx context.Context, p *part.Part) error {
	if p == nil {
		return errors.New("part is required")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.partsByID[p.ID]
	if existing == nil {
		return repository.ErrNotFound
	}
	sku := strings.TrimSpace(strings.ToLower(p.SKU))
	if other := r.s.partsBySKU[sku]; other != nil && other.ID != p.ID {
		return repository.ErrConflict
	}
	cp := *p
	r.s.partsByID[cp.ID] = &cp
	r.s.partsBySKU[sku] = &cp
	return nil
}

func (r *PartRepository) Delete(ctx context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	existing := r.s.partsByID[id]
	if existing == nil {
		return repository.ErrNotFound
	}
	delete(r.s.partsBySKU, strings.ToLower(existing.SKU))
	delete(r.s.partsByID, id)
	return nil
}

func (r *PartRepository) FindByID(ctx context.Context, id string) (*part.Part, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	p := r.s.partsByID[id]
	if p == nil {
		return nil, repository.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *PartRepository) FindByIDs(ctx context.Context, ids []string) ([]part.Part, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	out := make([]part.Part, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		p := r.s.partsByID[id]
		if p == nil {
			continue
		}
		out = append(out, *p)
	}
	return out, nil
}

func (r *PartRepository) List(ctx context.Context, limit, offset int) ([]part.Part, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	items := make([]*part.Part, 0, len(r.s.partsByID))
	for _, v := range r.s.partsByID {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginate(items, limit, offset, func(p *part.Part) part.Part { return *p }), nil
}

func (r *PartRepository) AdjustStock(ctx context.Context, partID string, movementType part.StockMovementType, quantity int, notes *string, createdByUserID *string) (*part.Part, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be > 0")
	}
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	p := r.s.partsByID[partID]
	if p == nil {
		return nil, repository.ErrNotFound
	}
	var newQty int
	switch movementType {
	case part.StockMovementIn:
		newQty = p.StockQuantity + quantity
	case part.StockMovementOut:
		if p.StockQuantity < quantity {
			return nil, errors.New("insufficient stock")
		}
		newQty = p.StockQuantity - quantity
	case part.StockMovementAdjustment:
		newQty = quantity
	default:
		return nil, errors.New("invalid movement type")
	}
	p.StockQuantity = newQty
	p.UpdatedAt = time.Now().UTC()
	cp := *p
	return &cp, nil
}

// -------------------- Service Orders --------------------

type ServiceOrderRepository struct{ s *Store }

func NewServiceOrderRepository(s *Store) *ServiceOrderRepository {
	return &ServiceOrderRepository{s: s}
}

func (r *ServiceOrderRepository) List(ctx context.Context, limit, offset int, status *order.Status) ([]order.ServiceOrderSummary, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	items := make([]*order.ServiceOrder, 0, len(r.s.serviceOrdersByID))
	for _, v := range r.s.serviceOrdersByID {
		if status != nil && *status != "" && v.Status != *status {
			continue
		}
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].OpenedAt.After(items[j].OpenedAt) })

	out := make([]order.ServiceOrderSummary, 0, len(items))
	for _, it := range items {
		cl := r.s.clientsByID[it.ClientID]
		veh := r.s.vehiclesByID[it.VehicleID]
		sum := order.ServiceOrderSummary{
			ID:         it.ID,
			Code:       it.Code,
			Status:     it.Status,
			OpenedAt:   it.OpenedAt,
			ClientID:   it.ClientID,
			VehicleID:  it.VehicleID,
			ClientName: "",
			Plate:      "",
		}
		if cl != nil {
			sum.ClientName = cl.Name
		}
		if veh != nil {
			sum.Plate = veh.Plate
		}
		out = append(out, sum)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(out) {
		return []order.ServiceOrderSummary{}, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func (r *ServiceOrderRepository) FindByID(ctx context.Context, id string) (*order.ServiceOrder, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[id]
	if so == nil {
		return nil, repository.ErrNotFound
	}
	cp := *so
	return &cp, nil
}

func (r *ServiceOrderRepository) GetDetailByID(ctx context.Context, id string) (*order.ServiceOrderDetail, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[id]
	if so == nil {
		return nil, repository.ErrNotFound
	}
	latest := latestBudget(r.s.budgetsByServiceOrder[id])
	var latestBudgetCopy *order.Budget
	if latest != nil {
		cp := *latest
		latestBudgetCopy = &cp
	}

	services := []order.BudgetServiceItem{}
	parts := []order.BudgetPartItem{}
	if latest != nil {
		services = append(services, r.s.budgetServicesByBudget[latest.ID]...)
		parts = append(parts, r.s.budgetPartsByBudget[latest.ID]...)
	}

	hist := r.s.statusHistoryBySO[id]
	histOut := append([]order.StatusHistoryEntry{}, hist...)

	return &order.ServiceOrderDetail{
		ServiceOrder:   *so,
		LatestBudget:   latestBudgetCopy,
		BudgetServices: services,
		BudgetParts:    parts,
		StatusHistory:  histOut,
	}, nil
}

func (r *ServiceOrderRepository) AverageExecutionMinutes(ctx context.Context, from, to *time.Time) (float64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	var total float64
	var count float64
	for _, so := range r.s.serviceOrdersByID {
		if so.ExecutionStart == nil || so.FinishedAt == nil {
			continue
		}
		if from != nil && so.FinishedAt.Before(*from) {
			continue
		}
		if to != nil && so.FinishedAt.After(*to) {
			continue
		}
		mins := so.FinishedAt.Sub(*so.ExecutionStart).Minutes()
		total += mins
		count++
	}
	if count == 0 {
		return 0, nil
	}
	return total / count, nil
}

func (r *ServiceOrderRepository) AverageServiceExecutionMinutes(ctx context.Context, serviceID *string, from, to *time.Time) ([]repository.ServiceExecutionAverage, error) {
	// In-memory repo doesn't track per-service execution timestamps (started/completed per service).
	// Returning empty data keeps this repo lightweight for unit tests that don't cover this metric.
	return []repository.ServiceExecutionAverage{}, nil
}

// -------------------- Service Order Flow --------------------

type ServiceOrderFlowRepository struct {
	s     *Store
	parts *PartRepository
}

func NewServiceOrderFlowRepository(s *Store, parts *PartRepository) *ServiceOrderFlowRepository {
	return &ServiceOrderFlowRepository{s: s, parts: parts}
}

func (r *ServiceOrderFlowRepository) CreateDraft(ctx context.Context, p repository.CreateServiceOrderDraftParams) (*order.ServiceOrder, *order.Budget, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := p.ServiceOrder
	b := p.Budget

	now := time.Now().UTC()
	if so.ID == "" {
		so.ID = uuid.NewString()
	}
	so.CreatedAt = now
	so.UpdatedAt = now
	if so.OpenedAt.IsZero() {
		so.OpenedAt = now
	}
	if so.Code == "" {
		so.Code = fmt.Sprintf("OS-%s-%s", now.Format("20060102"), strings.ToUpper(uuid.NewString()[:6]))
	}
	if _, ok := r.s.serviceOrdersByCode[so.Code]; ok {
		return nil, nil, repository.ErrConflict
	}

	b.ID = nonEmptyID(b.ID)
	b.ServiceOrderID = so.ID
	b.Version = 1
	b.Status = order.BudgetStatusDraft
	b.CreatedAt = now
	b.UpdatedAt = now

	so.Status = order.StatusReceived

	r.s.serviceOrdersByID[so.ID] = &so
	r.s.serviceOrdersByCode[so.Code] = &so

	r.s.budgetsByID[b.ID] = &b
	r.s.budgetsByServiceOrder[so.ID] = append(r.s.budgetsByServiceOrder[so.ID], &b)
	r.s.budgetServicesByBudget[b.ID] = append([]order.BudgetServiceItem{}, p.BudgetServices...)
	r.s.budgetPartsByBudget[b.ID] = append([]order.BudgetPartItem{}, p.BudgetParts...)

	r.appendHistoryLocked(so.ID, nil, so.Status, nil, nil, now)

	soCopy := so
	bCopy := b
	return &soCopy, &bCopy, nil
}

func (r *ServiceOrderFlowRepository) CreateBudgetRevision(ctx context.Context, p repository.CreateBudgetRevisionParams) (*order.Budget, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[p.ServiceOrderID]
	if so == nil {
		return nil, repository.ErrNotFound
	}
	if so.Status != order.StatusInDiagnosis {
		return nil, repository.ErrConflict
	}

	now := time.Now().UTC()
	latest := latestBudget(r.s.budgetsByServiceOrder[p.ServiceOrderID])
	version := 1
	if latest != nil {
		version = latest.Version + 1
	}
	if version < 2 {
		version = 2
	}

	b := p.Budget
	b.ID = nonEmptyID(b.ID)
	b.ServiceOrderID = p.ServiceOrderID
	b.Version = version
	b.Status = order.BudgetStatusDraft
	b.CreatedAt = now
	b.UpdatedAt = now

	r.s.budgetsByID[b.ID] = &b
	r.s.budgetsByServiceOrder[p.ServiceOrderID] = append(r.s.budgetsByServiceOrder[p.ServiceOrderID], &b)
	r.s.budgetServicesByBudget[b.ID] = append([]order.BudgetServiceItem{}, p.BudgetServices...)
	r.s.budgetPartsByBudget[b.ID] = append([]order.BudgetPartItem{}, p.BudgetParts...)

	cp := b
	return &cp, nil
}

func (r *ServiceOrderFlowRepository) StartDiagnosis(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return r.transition(serviceOrderID, order.StatusInDiagnosis, changedByUserID, nil)
}

func (r *ServiceOrderFlowRepository) SendLatestBudget(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[serviceOrderID]
	if so == nil {
		return repository.ErrNotFound
	}
	b := latestBudget(r.s.budgetsByServiceOrder[serviceOrderID])
	if b == nil {
		return repository.ErrNotFound
	}
	if b.Status != order.BudgetStatusDraft {
		return errors.New("budget not in DRAFT")
	}
	now := time.Now().UTC()
	b.Status = order.BudgetStatusSent
	b.SentAt = &now
	b.UpdatedAt = now

	return r.transitionLocked(so, order.StatusWaitingApproval, changedByUserID, nil, nil, now)
}

func (r *ServiceOrderFlowRepository) Finish(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[serviceOrderID]
	if so == nil {
		return repository.ErrNotFound
	}
	now := time.Now().UTC()
	so.FinishedAt = &now
	return r.transitionLocked(so, order.StatusFinished, changedByUserID, nil, nil, now)
}

func (r *ServiceOrderFlowRepository) Deliver(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[serviceOrderID]
	if so == nil {
		return repository.ErrNotFound
	}
	now := time.Now().UTC()
	so.DeliveredAt = &now
	return r.transitionLocked(so, order.StatusDelivered, changedByUserID, nil, nil, now)
}

func (r *ServiceOrderFlowRepository) GetClientViewByCode(ctx context.Context, code string, documentNumber string) (*repository.ClientServiceOrderView, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	timePtrRFC3339 := func(t *time.Time) *string {
		if t == nil {
			return nil
		}
		s := t.UTC().Format(time.RFC3339)
		return &s
	}

	so := r.s.serviceOrdersByCode[code]
	if so == nil {
		return nil, repository.ErrNotFound
	}
	cl := r.s.clientsByID[so.ClientID]
	if cl == nil || cl.DocumentNumber != documentNumber {
		return nil, repository.ErrNotFound
	}
	v := r.s.vehiclesByID[so.VehicleID]
	if v == nil {
		return nil, repository.ErrNotFound
	}
	b := latestBudget(r.s.budgetsByServiceOrder[so.ID])
	if b == nil {
		return nil, repository.ErrNotFound
	}
	bs := r.s.budgetServicesByBudget[b.ID]
	bp := r.s.budgetPartsByBudget[b.ID]
	hist := r.s.statusHistoryBySO[so.ID]
	return &repository.ClientServiceOrderView{
		Code:              so.Code,
		Status:            so.Status,
		ClientID:          so.ClientID,
		VehicleID:         so.VehicleID,
		OpenedAt:          so.OpenedAt.Format(time.RFC3339),
		CustomerComplaint: so.CustomerComplaint,

		VehiclePlate:           v.Plate,
		VehicleBrand:           v.Brand,
		VehicleModel:           v.Model,
		VehicleManufactureYear: v.ManufactureYear,
		VehicleModelYear:       v.ModelYear,
		VehicleColor:           v.Color,

		BudgetID:              b.ID,
		BudgetVersion:         b.Version,
		BudgetStatus:          b.Status,
		BudgetTotalCents:      b.TotalAmountCents,
		BudgetSentAt:          timePtrRFC3339(b.SentAt),
		BudgetApprovedAt:      timePtrRFC3339(b.ApprovedAt),
		BudgetRejectedAt:      timePtrRFC3339(b.RejectedAt),
		BudgetApprovedByName:  b.ApprovedByName,
		BudgetRejectionReason: b.RejectionReason,

		BudgetServices: bs,
		BudgetParts:    bp,
		StatusHistory:  hist,
	}, nil
}

func (r *ServiceOrderFlowRepository) ApproveLatestBudgetByCode(ctx context.Context, code string, documentNumber string, approvedByName *string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByCode[code]
	if so == nil {
		return repository.ErrNotFound
	}
	cl := r.s.clientsByID[so.ClientID]
	if cl == nil || cl.DocumentNumber != documentNumber {
		return repository.ErrNotFound
	}
	b := latestBudget(r.s.budgetsByServiceOrder[so.ID])
	if b == nil {
		return repository.ErrNotFound
	}
	if b.Status != order.BudgetStatusSent {
		return errors.New("budget not in SENT")
	}

	now := time.Now().UTC()
	b.Status = order.BudgetStatusApproved
	b.ApprovedAt = &now
	b.ApprovedByName = approvedByName
	b.UpdatedAt = now

	for _, it := range r.s.budgetPartsByBudget[b.ID] {
		p := r.s.partsByID[it.PartID]
		if p == nil {
			return repository.ErrNotFound
		}
		if p.StockQuantity < it.Quantity {
			return fmt.Errorf("insufficient stock for part %s", p.ID)
		}
		p.StockQuantity -= it.Quantity
		p.UpdatedAt = now
	}

	so.ExecutionStart = &now
	return r.transitionLocked(so, order.StatusInProgress, nil, nil, nil, now)
}

func (r *ServiceOrderFlowRepository) RejectLatestBudgetByCode(ctx context.Context, code string, documentNumber string, rejectionReason string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByCode[code]
	if so == nil {
		return repository.ErrNotFound
	}
	cl := r.s.clientsByID[so.ClientID]
	if cl == nil || cl.DocumentNumber != documentNumber {
		return repository.ErrNotFound
	}
	b := latestBudget(r.s.budgetsByServiceOrder[so.ID])
	if b == nil {
		return repository.ErrNotFound
	}
	if b.Status != order.BudgetStatusSent {
		return errors.New("budget not in SENT")
	}

	now := time.Now().UTC()
	b.Status = order.BudgetStatusRejected
	b.RejectedAt = &now
	b.RejectionReason = &rejectionReason
	b.UpdatedAt = now

	reason := rejectionReason
	return r.transitionLocked(so, order.StatusInDiagnosis, nil, &reason, nil, now)
}

func (r *ServiceOrderFlowRepository) transition(serviceOrderID string, to order.Status, changedByUserID *string, reason *string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	so := r.s.serviceOrdersByID[serviceOrderID]
	if so == nil {
		return repository.ErrNotFound
	}
	return r.transitionLocked(so, to, changedByUserID, reason, nil, time.Now().UTC())
}

func (r *ServiceOrderFlowRepository) transitionLocked(so *order.ServiceOrder, to order.Status, changedByUserID *string, reason *string, updates map[string]any, now time.Time) error {
	if so.Status == to {
		return nil
	}
	if !allowedTransition(so.Status, to) {
		return fmt.Errorf("invalid status transition %s -> %s", so.Status, to)
	}
	from := so.Status
	so.Status = to
	so.UpdatedAt = now
	r.appendHistoryLocked(so.ID, &from, to, changedByUserID, reason, now)
	_ = updates
	return nil
}

func (r *ServiceOrderFlowRepository) appendHistoryLocked(soID string, from *order.Status, to order.Status, changedByUserID *string, reason *string, changedAt time.Time) {
	r.s.statusHistoryBySO[soID] = append(r.s.statusHistoryBySO[soID], order.StatusHistoryEntry{
		FromStatus:      from,
		ToStatus:        to,
		ChangedAt:       changedAt,
		ChangedByUserID: changedByUserID,
		Reason:          reason,
	})
}

func allowedTransition(from, to order.Status) bool {
	switch from {
	case order.StatusReceived:
		return to == order.StatusInDiagnosis || to == order.StatusWaitingApproval
	case order.StatusInDiagnosis:
		return to == order.StatusWaitingApproval
	case order.StatusWaitingApproval:
		return to == order.StatusInProgress || to == order.StatusInDiagnosis
	case order.StatusInProgress:
		return to == order.StatusFinished
	case order.StatusFinished:
		return to == order.StatusDelivered
	default:
		return false
	}
}

func latestBudget(list []*order.Budget) *order.Budget {
	if len(list) == 0 {
		return nil
	}
	latest := list[0]
	for _, b := range list[1:] {
		if b.Version > latest.Version {
			latest = b
		}
	}
	return latest
}

func nonEmptyID(id string) string {
	if strings.TrimSpace(id) != "" {
		return id
	}
	return uuid.NewString()
}

func paginate[T any, R any](items []T, limit, offset int, mapFn func(T) R) []R {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []R{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	out := make([]R, 0, end-offset)
	for _, it := range items[offset:end] {
		out = append(out, mapFn(it))
	}
	return out
}
