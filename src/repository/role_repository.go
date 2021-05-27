package repository

import (
	"auth-service/src/domain"
	"gorm.io/gorm"
)

type roleRepository struct {
	Conn *gorm.DB
}

type RoleRepository interface {
	Create(role *domain.Role) error
	GetByName(roleName string) (*domain.Role, error)
}

func NewRoleRepository(Conn *gorm.DB) RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) Create(role *domain.Role) error {
	return r.Conn.Create(role).Error
}

func (r *roleRepository) GetByName(roleName string) (*domain.Role, error) {
	var role *domain.Role
	err := r.Conn.Where("RoleName=?", roleName).First(&role).Error
	return role, err
}
