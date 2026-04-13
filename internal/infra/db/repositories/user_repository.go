package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
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

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	if u == nil {
		return errors.New("user is required")
	}

	row := userRow{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		DeletedAt:    u.DeletedAt,
	}

	err := r.db.WithContext(ctx).Create(&row).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return repository.ErrConflict
	}

	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("lower(email) = ?", strings.ToLower(email)).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return mapUser(row), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return mapUser(row), nil
}

func mapUser(row userRow) *user.User {
	return &user.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         user.Role(row.Role),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    row.DeletedAt,
	}
}
