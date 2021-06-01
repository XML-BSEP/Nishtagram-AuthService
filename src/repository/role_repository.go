package repository

import (
	"auth-service/src/domain"
	"context"
	"gorm.io/gorm"
)

type roleRepository struct {
	Conn *gorm.DB
}

type RoleRepository interface {
	Create(context context.Context, role *domain.Role) error
	GetByName(context context.Context, roleName string) (*domain.Role, error)
}

func NewRoleRepository(conn *gorm.DB) RoleRepository {
	return &roleRepository{Conn: conn}
}

func (r *roleRepository) Create(context context.Context, role *domain.Role) error {
	return r.Conn.Create(role).Error
}

func (r *roleRepository) GetByName(context context.Context, roleName string) (*domain.Role, error) {
	var role *domain.Role
	err := r.Conn.Where("RoleName=?", roleName).First(&role).Error
	return role, err
}
