package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/greenledger/services/user-auth/internal/models"
	"github.com/greenledger/shared/database"
	"github.com/greenledger/shared/logger"
	"gorm.io/gorm"
)

// RoleRepository handles role data operations
type RoleRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *database.PostgresDB, logger *logger.Logger) *RoleRepository {
	return &RoleRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	err := r.db.WithContext(ctx).Create(role).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create role", err,
			logger.String("role_name", role.Name))
		return fmt.Errorf("failed to create role: %w", err)
	}

	r.logger.LogInfo(ctx, "role created successfully",
		logger.String("role_id", role.ID.String()),
		logger.String("role_name", role.Name))

	return nil
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get role by ID", err,
			logger.String("role_id", id.String()))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "name = ?", name).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		r.logger.LogError(ctx, "failed to get role by name", err,
			logger.String("role_name", name))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// List retrieves all roles
func (r *RoleRepository) List(ctx context.Context) ([]*models.Role, error) {
	var roles []*models.Role
	
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Order("name").
		Find(&roles).Error
	
	if err != nil {
		r.logger.LogError(ctx, "failed to list roles", err)
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// Update updates a role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	err := r.db.WithContext(ctx).Save(role).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to update role", err,
			logger.String("role_id", role.ID.String()))
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.Role{}, "id = ?", id).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to delete role", err,
			logger.String("role_id", id.String()))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignPermission assigns a permission to a role
func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			return fmt.Errorf("role not found: %w", err)
		}

		var permission models.Permission
		if err := tx.First(&permission, "id = ?", permissionID).Error; err != nil {
			return fmt.Errorf("permission not found: %w", err)
		}

		if err := tx.Model(&role).Association("Permissions").Append(&permission); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}

		return nil
	})
}

// RemovePermission removes a permission from a role
func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithTransaction(ctx, func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			return fmt.Errorf("role not found: %w", err)
		}

		var permission models.Permission
		if err := tx.First(&permission, "id = ?", permissionID).Error; err != nil {
			return fmt.Errorf("permission not found: %w", err)
		}

		if err := tx.Model(&role).Association("Permissions").Delete(&permission); err != nil {
			return fmt.Errorf("failed to remove permission: %w", err)
		}

		return nil
	})
}

// PermissionRepository handles permission data operations
type PermissionRepository struct {
	db     *database.PostgresDB
	logger *logger.Logger
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *database.PostgresDB, logger *logger.Logger) *PermissionRepository {
	return &PermissionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	err := r.db.WithContext(ctx).Create(permission).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to create permission", err,
			logger.String("permission_name", permission.Name))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// GetByName retrieves a permission by name
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	
	err := r.db.WithContext(ctx).First(&permission, "name = ?", name).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// List retrieves all permissions
func (r *PermissionRepository) List(ctx context.Context) ([]*models.Permission, error) {
	var permissions []*models.Permission
	
	err := r.db.WithContext(ctx).
		Order("resource, action").
		Find(&permissions).Error
	
	if err != nil {
		r.logger.LogError(ctx, "failed to list permissions", err)
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}

// BulkCreate creates multiple permissions
func (r *PermissionRepository) BulkCreate(ctx context.Context, permissions []*models.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).CreateInBatches(permissions, 100).Error
	if err != nil {
		r.logger.LogError(ctx, "failed to bulk create permissions", err)
		return fmt.Errorf("failed to bulk create permissions: %w", err)
	}

	return nil
}
