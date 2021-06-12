package repository

import (
	"auth-service/domain"
	"context"
	logger "github.com/jelena-vlajkov/logger/logger"
	"gorm.io/gorm"
)

type roleRepository struct {
	Conn *gorm.DB
	logger *logger.Logger
}

type RoleRepository interface {
	Create(context context.Context, role *domain.Role) error
	GetByName(context context.Context, roleName string) (*domain.Role, error)
}

func NewRoleRepository(conn *gorm.DB, logger *logger.Logger) RoleRepository {
	return &roleRepository{Conn: conn, logger: logger}
}

func (r *roleRepository) Create(context context.Context, role *domain.Role) error {
	err := r.Conn.Create(role).Error
	if err != nil {
		r.logger.Logger.Errorf("error while creating role, error: %v\n", err)
	}
	return err
}

func (r *roleRepository) GetByName(context context.Context, roleName string) (*domain.Role, error) {
	var role *domain.Role
	err := r.Conn.Where("RoleName=?", roleName).First(&role).Error

	if err != nil {
		r.logger.Logger.Errorf("error while getting role by name, error: %v\n", err)
	}
	return role, err
}
