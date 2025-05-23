package ports

import (
	"context"
	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/models/entities"
)

// RolerRepository define las operaciones disponibles para la gestión de roles y permisos
type RolerRepository interface {
	// Operaciones de Roles
	GetRoleByID(ctx context.Context, id string) (*entities.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entities.Role, error)
	UpdateRole(ctx context.Context, role *entities.Role) error
	IsRoleActive(ctx context.Context, id string) (bool, error)
	DeactivateRole(ctx context.Context, id string) error
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context) ([]entities.Role, error)
	IsRoleExist(ctx context.Context, param string) (bool, error)
	GetRoleByIDOrName(ctx context.Context, param string) (*entities.Role, error)

	// Operaciones de Permisos
	CreatePermission(ctx context.Context, permission *entities.Permission) error
	GetPermissionByID(ctx context.Context, id string) (*entities.Permission, error)
	UpdatePermission(ctx context.Context, permission *entities.Permission) error
	DeletePermission(ctx context.Context, id string) error
	ListPermissions(ctx context.Context) ([]entities.Permission, error)

	// Operaciones de Asignación
	AssignPermissionToRole(ctx context.Context, roleID string, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]entities.Permission, error)
}
