package order

import "time"

type Status string

const (
	StatusReceived        Status = "RECEIVED"
	StatusInDiagnosis     Status = "IN_DIAGNOSIS"
	StatusWaitingApproval Status = "WAITING_APPROVAL"
	StatusInProgress      Status = "IN_PROGRESS"
	StatusFinished        Status = "FINISHED"
	StatusDelivered       Status = "DELIVERED"
	StatusCanceled        Status = "CANCELED"
)

type ServiceOrder struct {
	ID                string
	Code              string
	ClientID          string
	VehicleID         string
	AssignedUserID    *string
	Status            Status
	CustomerComplaint *string
	OpenedAt          time.Time
	ExecutionStart    *time.Time
	FinishedAt        *time.Time
	DeliveredAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type ServiceOrderSummary struct {
	ID         string
	Code       string
	Status     Status
	OpenedAt   time.Time
	ClientID   string
	ClientName string
	VehicleID  string
	Plate      string
}
