package repositories

import (
	"context"
	"domain/delivery/models/roles"
	"domain/delivery/models/users"
	"domain/delivery/ports"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create inserta un nuevo usuario y su perfil si existe
func (r *userRepository) Create(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		if user.Profile != nil {
			user.Profile.UserID = user.ID
			if err := tx.Create(user.Profile).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByID obtiene un usuario por ID incluyendo su perfil y roles activos
func (r *userRepository) GetByID(ctx context.Context, id string) (*users.User, error) {
	var usr users.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Roles", "is_active = ?", true).
		Preload("Roles.Role").
		Preload("Sessions", "expires_at > NOW()").
		First(&usr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &usr, nil
}

// GetByEmail obtiene un usuario por email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	var usr users.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Roles", "is_active = ?", true).
		Preload("Roles.Role").
		Preload("Sessions", "expires_at > NOW()").
		Where("email = ?", email).
		First(&usr).Error
	if err != nil {
		return nil, err
	}
	return &usr, nil
}

// Update actualiza la información del usuario
func (r *userRepository) Update(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete realiza un soft delete del usuario
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&users.User{}, "id = ?", id).Error
}

// GetProfileByUserID obtiene el perfil de un usuario
func (r *userRepository) GetProfileByUserID(ctx context.Context, userID string) (*users.Profile, error) {
	var profile users.Profile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile actualiza el perfil del usuario
func (r *userRepository) UpdateProfile(ctx context.Context, profile *users.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// CreateSession crea una nueva sesión
func (r *userRepository) CreateSession(ctx context.Context, session *users.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSessionByToken obtiene una sesión por su token
func (r *userRepository) GetSessionByToken(ctx context.Context, token string) (*users.UserSession, error) {
	var session users.UserSession
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > NOW()", token).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetActiveSessionsByUserID obtiene todas las sesiones activas de un usuario
func (r *userRepository) GetActiveSessionsByUserID(ctx context.Context, userID string) ([]users.UserSession, error) {
	var sessions []users.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > NOW()", userID).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// DeleteSession elimina una sesión específica
func (r *userRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Delete(&users.UserSession{}, "id = ?", sessionID).Error
}

// CleanExpiredSessions elimina todas las sesiones expiradas
func (r *userRepository) CleanExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&users.UserSession{}, "expires_at <= NOW()").Error
}

// AssignRoleToUser asigna un rol a un usuario
func (r *userRepository) AssignRoleToUser(ctx context.Context, userID string, roleID string, assignedBy string) error {
	userRole := users.Role{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
	}
	return r.db.WithContext(ctx).Create(&userRole).Error
}

// RemoveRoleFromUser remueve un rol de un usuario
func (r *userRepository) RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error {
	return r.db.WithContext(ctx).
		Model(&users.Role{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Update("is_active", false).Error
}

// GetUserRoles obtiene todos los roles de un usuario
func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]roles.Role, error) {
	var roles []roles.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND user_roles.is_active = ?", userID, true).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetUserPermissions obtiene todos los permisos de un usuario a través de sus roles
func (r *userRepository) GetUserPermissions(ctx context.Context, userID string) ([]roles.Permission, error) {
	var permissions []roles.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND user_roles.is_active = ?", userID, true).
		Distinct().
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// MarkEmailAsVerified marca el email como verificado
func (r *userRepository) MarkEmailAsVerified(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&users.User{}).
		Where("id = ?", userID).
		Update("email_verified_at", r.db.NowFunc()).Error
}

// MarkPhoneAsVerified marca el teléfono como verificado
func (r *userRepository) MarkPhoneAsVerified(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&users.User{}).
		Where("id = ?", userID).
		Update("phone_verified_at", r.db.NowFunc()).Error
}
