package services

import (
	"errors"
	"time"

	"hr-system/internal/models"
	"hr-system/internal/repository"
	"hr-system/pkg/utils"

	"github.com/google/uuid"
)

type UserService struct {
	repo     *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewUserService(repo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{repo: repo, roleRepo: roleRepo}
}

func (s *UserService) Register(user *models.User) error {
	exists, err := s.repo.EmailExists(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashed

	// Default to employee role if none specified
	if user.RoleID == nil {
		role, err := s.roleRepo.GetByName(models.RoleEmployee)
		if err == nil {
			user.RoleID = &role.RoleID
		}
	}

	user.IsActive = true
	return s.repo.Create(user)
}

func (s *UserService) Login(email, password string) (map[string]interface{}, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}
	if err := utils.ComparePasswords(user.Password, password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(user.Email, user.UserID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token":      token,
		"expires_at": time.Now().Add(24 * time.Hour),
		"user": map[string]interface{}{
			"user_id":  user.UserID,
			"email":    user.Email,
			"role":     user.Role,
			"is_active": user.IsActive,
		},
	}, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *UserService) UpdateUser(updates *models.User) (*models.User, error) {
	existing, err := s.repo.GetUserByID(updates.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if updates.Email != "" && updates.Email != existing.Email {
		exists, err := s.repo.EmailExists(updates.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already in use")
		}
		existing.Email = updates.Email
	}

	if updates.RoleID != nil {
		existing.RoleID = updates.RoleID
	}
	existing.IsActive = updates.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(existing.UserID)
}

func (s *UserService) DeactivateUser(id uuid.UUID) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	user.IsActive = false
	return s.repo.Update(user)
}

func (s *UserService) SeedSuperAdmin(email, password string) error {
	exists, err := s.repo.EmailExists(email)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	role, err := s.roleRepo.GetByName(models.RoleSuperAdmin)
	if err != nil {
		return err
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:    email,
		Password: hashed,
		RoleID:   &role.RoleID,
		IsActive: true,
	}
	return s.repo.Create(user)
}
