package user

import "time"

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleManager   Role = "MANAGER"
	RoleMechanic  Role = "MECHANIC"
	RoleAttendant Role = "ATTENDANT"
	RoleViewer    Role = "VIEWER"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
