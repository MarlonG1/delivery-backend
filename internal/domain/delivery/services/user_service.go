package services

import (
	"context"
	"domain/delivery/models/users"
	"domain/delivery/ports"
	"shared/logs"
)

type userProfileService struct {
	userRepo ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) ports.UserService {
	return &userProfileService{
		userRepo: userRepo,
	}
}

func (s *userProfileService) GetUserInfo(ctx context.Context, userID string) (*users.User, error) {
	userLogged, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logs.Error("Failed to get users profile", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	return userLogged, nil
}
