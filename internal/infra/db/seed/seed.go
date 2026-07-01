package seed

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brazilian-utils/go/cnpj"
	"github.com/brazilian-utils/go/cpf"
	"github.com/brazilian-utils/go/licenseplate"
	"github.com/bxcodec/faker/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Options struct {
	Force      bool
	RandomSeed int64

	Users    int
	Clients  int
	Services int
	Parts    int
	Orders   int

	Now time.Time
}

func Run(ctx context.Context, gormDB *gorm.DB, opts Options) error {
	if gormDB == nil {
		return errors.New("gormDB is required")
	}

	opts = withDefaults(opts)
	if err := validate(opts); err != nil {
		return err
	}

	return gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if opts.Force {
			if err := truncateAll(tx); err != nil {
				return err
			}
		} else {
			seeded, err := hasSeedMarker(tx)
			if err != nil {
				return err
			}
			if seeded {
				log.Printf("Seed skipped: %s already exists", adminEmail)
				return nil
			}
		}

		users, err := seedUsers(tx, opts)
		if err != nil {
			return err
		}
		clients, vehicles, err := seedClientsAndVehicles(tx, opts)
		if err != nil {
			return err
		}
		services, err := seedServices(tx, opts)
		if err != nil {
			return err
		}
		parts, err := seedParts(tx, opts)
		if err != nil {
			return err
		}
		if err := seedOrders(tx, opts, users, clients, vehicles, services, parts); err != nil {
			return err
		}

		log.Printf(
			"Seed complete: users=%d clients=%d vehicles=%d services=%d parts=%d orders=%d",
			len(users), len(clients), len(vehicles), len(services), len(parts), opts.Orders,
		)
		log.Printf("Seed login: email=%s password=%s", adminEmail, adminPassword)
		return nil
	})
}

const (
	adminEmail    = "admin@tech.local"
	adminPassword = "Senha@123"
)

func hasSeedMarker(tx *gorm.DB) (bool, error) {
	var count int64
	err := tx.
		Table("users").
		Where("deleted_at IS NULL").
		Where("lower(email) = lower(?)", adminEmail).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func truncateAll(tx *gorm.DB) error {
	return tx.Exec(`
		TRUNCATE TABLE
			stock_movements,
			service_order_parts,
			service_order_services,
			budget_parts,
			budget_services,
			budgets,
			service_order_status_history,
			service_orders,
			vehicles,
			clients,
			parts,
			services,
			users
		RESTART IDENTITY CASCADE;
	`).Error
}

func withDefaults(opts Options) Options {
	if opts.RandomSeed == 0 {
		opts.RandomSeed = 123
	}
	if opts.Users == 0 {
		opts.Users = 8
	}
	if opts.Clients == 0 {
		opts.Clients = 12
	}
	if opts.Services == 0 {
		opts.Services = 8
	}
	if opts.Parts == 0 {
		opts.Parts = 24
	}
	if opts.Orders == 0 {
		opts.Orders = 12
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	return opts
}

func validate(opts Options) error {
	if opts.Users < 1 {
		return fmt.Errorf("SEED_USERS must be >= 1")
	}
	if opts.Clients < 1 {
		return fmt.Errorf("SEED_CLIENTS must be >= 1")
	}
	if opts.Services < 1 {
		return fmt.Errorf("SEED_SERVICES must be >= 1")
	}
	if opts.Parts < 1 {
		return fmt.Errorf("SEED_PARTS must be >= 1")
	}
	if opts.Orders < 1 {
		return fmt.Errorf("SEED_ORDERS must be >= 1")
	}
	return nil
}

type seededUser struct {
	ID   string
	Role string
}

func seedUsers(tx *gorm.DB, opts Options) ([]seededUser, error) {
	pwHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	rows := make([]userRow, 0, opts.Users)
	rows = append(rows, userRow{
		ID:           uuid.NewString(),
		Name:         "Admin",
		Email:        adminEmail,
		PasswordHash: string(pwHash),
		Role:         "ADMIN",
		CreatedAt:    opts.Now,
		UpdatedAt:    opts.Now,
	})

	roles := []string{"MANAGER", "MECHANIC", "ATTENDANT", "VIEWER"}
	for i := 1; i < opts.Users; i++ {
		rows = append(rows, userRow{
			ID:           uuid.NewString(),
			Name:         faker.Name(),
			Email:        fmt.Sprintf("user%02d@tech.local", i),
			PasswordHash: string(pwHash),
			Role:         roles[(i-1)%len(roles)],
			CreatedAt:    opts.Now,
			UpdatedAt:    opts.Now,
		})
	}

	if err := tx.Create(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]seededUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, seededUser{ID: r.ID, Role: r.Role})
	}
	return out, nil
}

type seededClient struct {
	ID             string
	DocumentType   string
	DocumentNumber string
}

type seededVehicle struct {
	ID       string
	ClientID string
	Plate    string
}

func seedClientsAndVehicles(tx *gorm.DB, opts Options) ([]seededClient, []seededVehicle, error) {
	docUsed := map[string]bool{}
	plateUsed := map[string]bool{}
	chassisUsed := map[string]bool{}

	clients := make([]clientRow, 0, opts.Clients)
	vehicles := make([]vehicleRow, 0, opts.Clients)

	for i := 0; i < opts.Clients; i++ {
		clID := uuid.NewString()

		docType := "CPF"
		if i%4 == 0 {
			docType = "CNPJ"
		}

		docNumber := uniqueDocNumber(docType, docUsed)
		name := faker.Name()
		email := fmt.Sprintf("client%02d@tech.local", i+1)
		phone := fmt.Sprintf("+55 11 9%08d", 10000000+i)

		clients = append(clients, clientRow{
			ID:             clID,
			DocumentType:   docType,
			DocumentNumber: docNumber,
			Name:           name,
			Email:          &email,
			Phone:          &phone,
			CreatedAt:      opts.Now,
			UpdatedAt:      opts.Now,
		})

		plate := uniquePlate(plateUsed)
		modelYear := 2020 + (i % 6)
		brand := []string{"Fiat", "Volkswagen", "Chevrolet", "Ford", "Toyota", "Honda"}[i%6]
		model := []string{"Uno", "Gol", "Onix", "Ka", "Corolla", "Civic"}[i%6]
		color := []string{"Preto", "Branco", "Prata", "Vermelho", "Azul"}[i%5]
		mileage := 10000 + (i * 1234)
		chassis := uniqueChassis(chassisUsed)

		vehicles = append(vehicles, vehicleRow{
			ID:              uuid.NewString(),
			ClientID:        clID,
			Plate:           plate,
			Brand:           brand,
			Model:           model,
			ManufactureYear: ptrInt(modelYear - 1),
			ModelYear:       modelYear,
			Color:           &color,
			Mileage:         &mileage,
			Chassis:         &chassis,
			CreatedAt:       opts.Now,
			UpdatedAt:       opts.Now,
		})
	}

	if err := tx.Create(&clients).Error; err != nil {
		return nil, nil, err
	}
	if err := tx.Create(&vehicles).Error; err != nil {
		return nil, nil, err
	}

	outClients := make([]seededClient, 0, len(clients))
	for _, c := range clients {
		outClients = append(outClients, seededClient{ID: c.ID, DocumentType: c.DocumentType, DocumentNumber: c.DocumentNumber})
	}
	outVehicles := make([]seededVehicle, 0, len(vehicles))
	for _, v := range vehicles {
		outVehicles = append(outVehicles, seededVehicle{ID: v.ID, ClientID: v.ClientID, Plate: v.Plate})
	}
	return outClients, outVehicles, nil
}

type seededService struct {
	ID          string
	Name        string
	BasePrice   int64
	Minutes     int
	Description *string
}

func seedServices(tx *gorm.DB, opts Options) ([]seededService, error) {
	catalog := []struct {
		name    string
		minutes int
		price   int64
	}{
		{"Troca de óleo", 45, 15900},
		{"Alinhamento e balanceamento", 60, 19900},
		{"Troca de pastilhas de freio", 90, 34900},
		{"Revisão completa", 180, 79900},
		{"Diagnóstico eletrônico", 60, 24900},
		{"Troca de bateria", 30, 29900},
		{"Troca de filtro de ar", 20, 7900},
		{"Higienização do ar-condicionado", 50, 17900},
		{"Troca de velas", 60, 22900},
		{"Troca de amortecedores", 120, 89900},
	}

	n := min(opts.Services, len(catalog))
	rows := make([]serviceRow, 0, n)
	out := make([]seededService, 0, n)
	for i := 0; i < n; i++ {
		desc := fmt.Sprintf("%s (%s)", catalog[i].name, strings.ToLower(faker.Sentence()))
		row := serviceRow{
			ID:               uuid.NewString(),
			Name:             catalog[i].name,
			Description:      &desc,
			BasePriceCents:   catalog[i].price,
			EstimatedMinutes: catalog[i].minutes,
			Active:           true,
			CreatedAt:        opts.Now,
			UpdatedAt:        opts.Now,
		}
		rows = append(rows, row)
		out = append(out, seededService{
			ID:          row.ID,
			Name:        row.Name,
			BasePrice:   row.BasePriceCents,
			Minutes:     row.EstimatedMinutes,
			Description: row.Description,
		})
	}

	if err := tx.Create(&rows).Error; err != nil {
		return nil, err
	}
	return out, nil
}

type seededPart struct {
	ID        string
	Name      string
	UnitPrice int64
}

func seedParts(tx *gorm.DB, opts Options) ([]seededPart, error) {
	rows := make([]partRow, 0, opts.Parts)
	out := make([]seededPart, 0, opts.Parts)

	for i := 0; i < opts.Parts; i++ {
		desc := faker.Sentence()
		price := int64(3500 + (i%12)*1200)
		stock := 10 + (i % 15)
		row := partRow{
			ID:             uuid.NewString(),
			SKU:            fmt.Sprintf("SKU-%04d", i+1),
			Name:           fmt.Sprintf("Peça %02d - %s", i+1, faker.Word()),
			Description:    &desc,
			UnitPriceCents: price,
			StockQuantity:  stock,
			Active:         true,
			CreatedAt:      opts.Now,
			UpdatedAt:      opts.Now,
		}
		rows = append(rows, row)
		out = append(out, seededPart{ID: row.ID, Name: row.Name, UnitPrice: row.UnitPriceCents})
	}

	if err := tx.Create(&rows).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func seedOrders(
	tx *gorm.DB,
	opts Options,
	users []seededUser,
	clients []seededClient,
	vehicles []seededVehicle,
	services []seededService,
	parts []seededPart,
) error {
	if len(clients) == 0 || len(vehicles) == 0 || len(services) == 0 {
		return errors.New("missing seed data for orders")
	}

	mechanicID := pickUserID(users, "MECHANIC")
	adminID := pickUserID(users, "ADMIN")

	vehicleByClient := map[string][]seededVehicle{}
	for _, v := range vehicles {
		vehicleByClient[v.ClientID] = append(vehicleByClient[v.ClientID], v)
	}

	for i := 0; i < opts.Orders; i++ {
		cl := clients[i%len(clients)]
		vs := vehicleByClient[cl.ID]
		if len(vs) == 0 {
			continue
		}
		v := vs[i%len(vs)]

		stage := pickOrderStage(i)
		openedAt := opts.Now.Add(-time.Duration((i%30)+1) * 6 * time.Hour)
		diagAt := openedAt.Add(30 * time.Minute)
		waitApprovalAt := diagAt.Add(45 * time.Minute)
		execAt := waitApprovalAt.Add(2 * time.Hour)
		finishedAt := execAt.Add(time.Duration(45+(i%120)) * time.Minute)
		deliveredAt := finishedAt.Add(2 * time.Hour)

		soID := uuid.NewString()
		code := fmt.Sprintf("SO-%04d", i+1)
		so := serviceOrderRow{
			ID:        soID,
			Code:      code,
			ClientID:  cl.ID,
			VehicleID: v.ID,
			Status:    stage.Status,
			OpenedAt:  openedAt,
			CreatedAt: openedAt,
			UpdatedAt: openedAt,
		}
		if mechanicID != "" && stage.AssignMechanic {
			so.AssignedUserID = &mechanicID
		}
		if stage.HasDiagnosis {
			so.DiagnosisStartedAt = &diagAt
		}
		if stage.HasWaitingApproval {
			so.WaitingApprovalAt = &waitApprovalAt
		}
		if stage.HasExecution {
			so.ExecutionStartedAt = &execAt
		}
		if stage.HasFinished {
			so.FinishedAt = &finishedAt
		}
		if stage.HasDelivered {
			so.DeliveredAt = &deliveredAt
		}
		if err := tx.Create(&so).Error; err != nil {
			return err
		}

		hist := buildStatusHistory(soID, adminID, openedAt, diagAt, waitApprovalAt, execAt, finishedAt, deliveredAt, stage)
		if len(hist) > 0 {
			if err := tx.Create(&hist).Error; err != nil {
				return err
			}
		}

		budgetID := uuid.NewString()
		linesSvc, totalSvc := buildBudgetServices(budgetID, opts.Now, services, i)
		linesPart, totalPart := buildBudgetParts(budgetID, opts.Now, parts, i)
		total := totalSvc + totalPart

		budgetStatus, sentAt, decidedAt, approvedAt := budgetStatusForStage(stage, waitApprovalAt, execAt)
		b := budgetRow{
			ID:               budgetID,
			ServiceOrderID:   soID,
			Version:          1,
			Status:           budgetStatus,
			TotalAmountCents: total,
			SentAt:           sentAt,
			DecidedAt:        decidedAt,
			ApprovedAt:       approvedAt,
			CreatedAt:        openedAt,
			UpdatedAt:        openedAt,
		}
		if err := tx.Create(&b).Error; err != nil {
			return err
		}
		if len(linesSvc) > 0 {
			if err := tx.Create(&linesSvc).Error; err != nil {
				return err
			}
		}
		if len(linesPart) > 0 {
			if err := tx.Create(&linesPart).Error; err != nil {
				return err
			}
		}

		if err := seedOrderItemsFromBudget(tx, opts, soID, mechanicID, stage, execAt, finishedAt, linesSvc, linesPart); err != nil {
			return err
		}
	}

	return nil
}

func seedOrderItemsFromBudget(
	tx *gorm.DB,
	opts Options,
	serviceOrderID string,
	mechanicID string,
	stage orderStage,
	execAt time.Time,
	finishedAt time.Time,
	budgetServices []budgetServiceRow,
	budgetParts []budgetPartRow,
) error {
	var itemStatus string
	var startedAt *time.Time
	var completedAt *time.Time
	switch stage.Status {
	case "IN_PROGRESS":
		itemStatus = "IN_PROGRESS"
		startedAt = &execAt
	case "FINISHED", "DELIVERED":
		itemStatus = "COMPLETED"
		startedAt = &execAt
		completedAt = &finishedAt
	default:
		itemStatus = "PENDING"
	}

	for _, bs := range budgetServices {
		row := serviceOrderServiceRow{
			ID:              uuid.NewString(),
			ServiceOrderID:  serviceOrderID,
			ServiceID:       bs.ServiceID,
			BudgetServiceID: &bs.ID,
			Description:     bs.Description,
			Quantity:        bs.Quantity,
			UnitPriceCents:  bs.UnitPriceCents,
			TotalPriceCents: bs.TotalPriceCents,
			Status:          itemStatus,
			StartedAt:       startedAt,
			CompletedAt:     completedAt,
			CreatedAt:       opts.Now,
			UpdatedAt:       opts.Now,
		}
		if mechanicID != "" {
			row.AssignedUserID = &mechanicID
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	for _, bp := range budgetParts {
		row := serviceOrderPartRow{
			ID:              uuid.NewString(),
			ServiceOrderID:  serviceOrderID,
			PartID:          bp.PartID,
			BudgetPartID:    &bp.ID,
			Description:     bp.Description,
			Quantity:        bp.Quantity,
			UnitPriceCents:  bp.UnitPriceCents,
			TotalPriceCents: bp.TotalPriceCents,
			Status:          itemStatus,
			CreatedAt:       opts.Now,
			UpdatedAt:       opts.Now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

type orderStage struct {
	Status      string
	Transitions []string

	AssignMechanic     bool
	HasDiagnosis       bool
	HasWaitingApproval bool
	HasExecution       bool
	HasFinished        bool
	HasDelivered       bool
}

func pickOrderStage(i int) orderStage {
	// Deterministic distribution: enough FINISHED/DELIVERED rows to power metrics.
	switch i % 9 {
	case 0:
		return orderStage{Status: "RECEIVED", Transitions: []string{"RECEIVED"}}
	case 1:
		return orderStage{Status: "IN_DIAGNOSIS", Transitions: []string{"RECEIVED", "IN_DIAGNOSIS"}, HasDiagnosis: true}
	case 2:
		return orderStage{
			Status:             "WAITING_APPROVAL",
			Transitions:        []string{"RECEIVED", "IN_DIAGNOSIS", "WAITING_APPROVAL"},
			HasDiagnosis:       true,
			HasWaitingApproval: true,
		}
	case 3:
		return orderStage{
			Status:             "IN_PROGRESS",
			Transitions:        []string{"RECEIVED", "IN_DIAGNOSIS", "WAITING_APPROVAL", "IN_PROGRESS"},
			AssignMechanic:     true,
			HasDiagnosis:       true,
			HasWaitingApproval: true,
			HasExecution:       true,
		}
	case 4, 5:
		return orderStage{
			Status:             "FINISHED",
			Transitions:        []string{"RECEIVED", "IN_DIAGNOSIS", "WAITING_APPROVAL", "IN_PROGRESS", "FINISHED"},
			AssignMechanic:     true,
			HasDiagnosis:       true,
			HasWaitingApproval: true,
			HasExecution:       true,
			HasFinished:        true,
		}
	case 6:
		return orderStage{
			Status:             "DELIVERED",
			Transitions:        []string{"RECEIVED", "IN_DIAGNOSIS", "WAITING_APPROVAL", "IN_PROGRESS", "FINISHED", "DELIVERED"},
			AssignMechanic:     true,
			HasDiagnosis:       true,
			HasWaitingApproval: true,
			HasExecution:       true,
			HasFinished:        true,
			HasDelivered:       true,
		}
	default:
		return orderStage{
			Status:       "CANCELED",
			Transitions:  []string{"RECEIVED", "IN_DIAGNOSIS", "CANCELED"},
			HasDiagnosis: true,
		}
	}
}

func buildStatusHistory(
	serviceOrderID string,
	adminID string,
	openedAt time.Time,
	diagAt time.Time,
	waitApprovalAt time.Time,
	execAt time.Time,
	finishedAt time.Time,
	deliveredAt time.Time,
	stage orderStage,
) []serviceOrderStatusHistoryRow {
	if len(stage.Transitions) == 0 {
		return nil
	}

	changedAtFor := func(status string) time.Time {
		switch status {
		case "RECEIVED":
			return openedAt
		case "IN_DIAGNOSIS":
			return diagAt
		case "WAITING_APPROVAL":
			return waitApprovalAt
		case "IN_PROGRESS":
			return execAt
		case "FINISHED":
			return finishedAt
		case "DELIVERED":
			return deliveredAt
		case "CANCELED":
			return diagAt.Add(20 * time.Minute)
		default:
			return openedAt
		}
	}

	rows := make([]serviceOrderStatusHistoryRow, 0, len(stage.Transitions))
	var prev string
	for _, to := range stage.Transitions {
		var fromPtr *string
		if prev != "" {
			from := prev
			fromPtr = &from
		}
		rows = append(rows, serviceOrderStatusHistoryRow{
			ID:              uuid.NewString(),
			ServiceOrderID:  serviceOrderID,
			FromStatus:      fromPtr,
			ToStatus:        to,
			ChangedByUserID: nullableString(adminID),
			ChangedAt:       changedAtFor(to),
			CreatedAt:       openedAt,
		})
		prev = to
	}
	return rows
}

func budgetStatusForStage(stage orderStage, sentAt time.Time, approvedAt time.Time) (string, *time.Time, *time.Time, *time.Time) {
	switch stage.Status {
	case "WAITING_APPROVAL":
		return "SENT", &sentAt, nil, nil
	case "IN_PROGRESS", "FINISHED", "DELIVERED":
		return "APPROVED", &sentAt, &approvedAt, &approvedAt
	case "CANCELED":
		return "CANCELED", nil, nil, nil
	default:
		return "DRAFT", nil, nil, nil
	}
}

func buildBudgetServices(budgetID string, optsNow time.Time, services []seededService, i int) ([]budgetServiceRow, int64) {
	count := 1 + (i % 3)
	if count > len(services) {
		count = len(services)
	}

	var total int64
	rows := make([]budgetServiceRow, 0, count)
	for j := 0; j < count; j++ {
		svc := services[(i+j)%len(services)]
		serviceID := svc.ID
		qty := 1
		unit := svc.BasePrice
		lineTotal := unit * int64(qty)
		total += lineTotal

		rows = append(rows, budgetServiceRow{
			ID:              uuid.NewString(),
			BudgetID:        budgetID,
			ServiceID:       serviceID,
			Description:     svc.Name,
			Quantity:        qty,
			UnitPriceCents:  unit,
			TotalPriceCents: lineTotal,
			CreatedAt:       optsNow,
			UpdatedAt:       optsNow,
		})
	}
	return rows, total
}

func buildBudgetParts(budgetID string, optsNow time.Time, parts []seededPart, i int) ([]budgetPartRow, int64) {
	if len(parts) == 0 {
		return nil, 0
	}

	count := i % 3
	if count > len(parts) {
		count = len(parts)
	}

	var total int64
	rows := make([]budgetPartRow, 0, count)
	for j := 0; j < count; j++ {
		part := parts[(i*3+j)%len(parts)]
		partID := part.ID
		qty := 1 + (j % 2)
		unit := part.UnitPrice
		lineTotal := unit * int64(qty)
		total += lineTotal

		rows = append(rows, budgetPartRow{
			ID:              uuid.NewString(),
			BudgetID:        budgetID,
			PartID:          partID,
			Description:     part.Name,
			Quantity:        qty,
			UnitPriceCents:  unit,
			TotalPriceCents: lineTotal,
			CreatedAt:       optsNow,
			UpdatedAt:       optsNow,
		})
	}
	return rows, total
}

func pickUserID(users []seededUser, role string) string {
	for _, u := range users {
		if u.Role == role {
			return u.ID
		}
	}
	return ""
}

func uniqueDocNumber(docType string, used map[string]bool) string {
	for {
		var doc string
		if docType == "CNPJ" {
			doc = cnpj.Generate(0)
		} else {
			doc = cpf.Generate()
		}
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		if used[doc] {
			continue
		}
		used[doc] = true
		return doc
	}
}

func uniquePlate(used map[string]bool) string {
	for {
		plate := licenseplate.Generate("LLLNLNN")
		plate = strings.ToUpper(strings.TrimSpace(plate))
		if plate == "" || used[plate] {
			continue
		}
		if !licenseplate.IsValid(plate, "") {
			continue
		}
		used[plate] = true
		return plate
	}
}

func uniqueChassis(used map[string]bool) string {
	for {
		v := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
		if len(v) > 40 {
			v = v[:40]
		}
		if used[v] {
			continue
		}
		used[v] = true
		return v
	}
}

func ptrInt(v int) *int { return &v }

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type userRow struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey"`
	Name         string     `gorm:"column:name"`
	Email        string     `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	Role         string     `gorm:"column:role"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (userRow) TableName() string { return "users" }

type clientRow struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey"`
	DocumentType   string     `gorm:"column:document_type"`
	DocumentNumber string     `gorm:"column:document_number"`
	Name           string     `gorm:"column:name"`
	Email          *string    `gorm:"column:email"`
	Phone          *string    `gorm:"column:phone"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (clientRow) TableName() string { return "clients" }

type vehicleRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ClientID        string     `gorm:"column:client_id;type:uuid"`
	Plate           string     `gorm:"column:plate"`
	Brand           string     `gorm:"column:brand"`
	Model           string     `gorm:"column:model"`
	ManufactureYear *int       `gorm:"column:manufacture_year"`
	ModelYear       int        `gorm:"column:model_year"`
	Color           *string    `gorm:"column:color"`
	Mileage         *int       `gorm:"column:mileage"`
	Chassis         *string    `gorm:"column:chassis"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (vehicleRow) TableName() string { return "vehicles" }

type serviceRow struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	Name             string     `gorm:"column:name"`
	Description      *string    `gorm:"column:description"`
	BasePriceCents   int64      `gorm:"column:base_price_cents"`
	EstimatedMinutes int        `gorm:"column:estimated_minutes"`
	Active           bool       `gorm:"column:active"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (serviceRow) TableName() string { return "services" }

type partRow struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey"`
	SKU            string     `gorm:"column:sku"`
	Name           string     `gorm:"column:name"`
	Description    *string    `gorm:"column:description"`
	UnitPriceCents int64      `gorm:"column:unit_price_cents"`
	StockQuantity  int        `gorm:"column:stock_quantity"`
	Active         bool       `gorm:"column:active"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (partRow) TableName() string { return "parts" }

type serviceOrderRow struct {
	ID                 string     `gorm:"column:id;type:uuid;primaryKey"`
	Code               string     `gorm:"column:code"`
	ClientID           string     `gorm:"column:client_id;type:uuid"`
	VehicleID          string     `gorm:"column:vehicle_id;type:uuid"`
	AssignedUserID     *string    `gorm:"column:assigned_user_id;type:uuid"`
	Status             string     `gorm:"column:status"`
	OpenedAt           time.Time  `gorm:"column:opened_at"`
	DiagnosisStartedAt *time.Time `gorm:"column:diagnosis_started_at"`
	WaitingApprovalAt  *time.Time `gorm:"column:waiting_approval_at"`
	ExecutionStartedAt *time.Time `gorm:"column:execution_started_at"`
	FinishedAt         *time.Time `gorm:"column:finished_at"`
	DeliveredAt        *time.Time `gorm:"column:delivered_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderRow) TableName() string { return "service_orders" }

type serviceOrderStatusHistoryRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id;type:uuid"`
	FromStatus      *string    `gorm:"column:from_status"`
	ToStatus        string     `gorm:"column:to_status"`
	ChangedByUserID *string    `gorm:"column:changed_by_user_id;type:uuid"`
	Reason          *string    `gorm:"column:reason"`
	ChangedAt       time.Time  `gorm:"column:changed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderStatusHistoryRow) TableName() string { return "service_order_status_history" }

type budgetRow struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID   string     `gorm:"column:service_order_id;type:uuid"`
	Version          int        `gorm:"column:version"`
	Status           string     `gorm:"column:status"`
	TotalAmountCents int64      `gorm:"column:total_amount_cents"`
	Notes            *string    `gorm:"column:notes"`
	SentAt           *time.Time `gorm:"column:sent_at"`
	DecidedAt        *time.Time `gorm:"column:decided_at"`
	ApprovedAt       *time.Time `gorm:"column:approved_at"`
	RejectedAt       *time.Time `gorm:"column:rejected_at"`
	ApprovedByName   *string    `gorm:"column:approved_by_name"`
	RejectionReason  *string    `gorm:"column:rejection_reason"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (budgetRow) TableName() string { return "budgets" }

type budgetServiceRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	BudgetID        string     `gorm:"column:budget_id;type:uuid"`
	ServiceID       string     `gorm:"column:service_id;type:uuid"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetServiceRow) TableName() string { return "budget_services" }

type budgetPartRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	BudgetID        string     `gorm:"column:budget_id;type:uuid"`
	PartID          string     `gorm:"column:part_id;type:uuid"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetPartRow) TableName() string { return "budget_parts" }

type serviceOrderServiceRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id;type:uuid"`
	ServiceID       string     `gorm:"column:service_id;type:uuid"`
	BudgetServiceID *string    `gorm:"column:budget_service_id;type:uuid"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	Status          string     `gorm:"column:status"`
	AssignedUserID  *string    `gorm:"column:assigned_user_id;type:uuid"`
	StartedAt       *time.Time `gorm:"column:started_at"`
	CompletedAt     *time.Time `gorm:"column:completed_at"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderServiceRow) TableName() string { return "service_order_services" }

type serviceOrderPartRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id;type:uuid"`
	PartID          string     `gorm:"column:part_id;type:uuid"`
	BudgetPartID    *string    `gorm:"column:budget_part_id;type:uuid"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	Status          string     `gorm:"column:status"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderPartRow) TableName() string { return "service_order_parts" }
