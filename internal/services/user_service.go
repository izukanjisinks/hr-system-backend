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
	repo                 *repository.UserRepository
	roleRepo             *repository.RoleRepository
	passwordPolicyService *PasswordPolicyService
}

func NewUserService(repo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{repo: repo, roleRepo: roleRepo}
}

// SetPasswordPolicyService sets the password policy service (called after initialization)
func (s *UserService) SetPasswordPolicyService(policyService *PasswordPolicyService) {
	s.passwordPolicyService = policyService
}

func (s *UserService) Register(user *models.User) error {
	exists, err := s.repo.EmailExists(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	// Validate password against policy if available
	if s.passwordPolicyService != nil {
		if err := s.passwordPolicyService.ValidateNewPassword(uuid.Nil, user.Password, ""); err != nil {
			return err
		}
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
	now := time.Now()
	user.PasswordChangedAt = &now

	// Set password expiry if policy requires it
	if s.passwordPolicyService != nil {
		user.PasswordExpiresAt = s.passwordPolicyService.CalculatePasswordExpiry()
	}

	if err := s.repo.Create(user); err != nil {
		return err
	}

	// Record password in history
	if s.passwordPolicyService != nil {
		_ = s.passwordPolicyService.RecordPasswordChange(user.UserID, hashed)
	}

	return nil
}

func (s *UserService) Login(email, password string) (map[string]interface{}, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	// Check account lockout status
	if s.passwordPolicyService != nil {
		locked, reason := s.passwordPolicyService.CheckAccountLockout(user)
		if locked {
			return nil, errors.New(reason)
		}
	}

	// Verify password
	if err := utils.ComparePasswords(user.Password, password); err != nil {
		// Failed login attempt - increment counter
		if s.passwordPolicyService != nil {
			user.FailedLoginAttempts++

			// Check if account should be locked
			if s.passwordPolicyService.ShouldLockAccount(user.FailedLoginAttempts) {
				lockoutTime := s.passwordPolicyService.CalculateLockoutTime()
				user.LockedUntil = &lockoutTime
			}

			// Update user with failed attempt info
			_ = s.repo.Update(user)
		}
		return nil, errors.New("invalid email or password")
	}

	// Successful login - reset failed attempts and unlock if temporarily locked
	if s.passwordPolicyService != nil {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		_ = s.repo.Update(user)
	}

	// Check password expiry
	var passwordExpired bool
	var passwordExpiringSoon bool
	var daysUntilExpiry int
	if s.passwordPolicyService != nil {
		passwordExpired, passwordExpiringSoon, daysUntilExpiry = s.passwordPolicyService.CheckPasswordExpiry(user)
		if passwordExpired {
			return nil, errors.New("password has expired, please change your password")
		}
	}

	// Generate token with appropriate expiry
	var tokenExpiry time.Duration
	if s.passwordPolicyService != nil {
		tokenExpiry = time.Duration(s.passwordPolicyService.GetSessionTimeout()) * time.Second
	} else {
		tokenExpiry = 24 * time.Hour // Default 24 hours
	}

	token, err := utils.GenerateToken(user.Email, user.UserID)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"token":      token,
		"expires_at": time.Now().Add(tokenExpiry),
		"user": map[string]interface{}{
			"user_id":         user.UserID,
			"email":           user.Email,
			"role":            user.Role,
			"is_active":       user.IsActive,
			"change_password": user.ChangePassword,
		},
	}

	// Add password expiry warnings if applicable
	if passwordExpiringSoon {
		response["password_warning"] = map[string]interface{}{
			"message":           "Your password is expiring soon",
			"days_until_expiry": daysUntilExpiry,
		}
	}

	return response, nil
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

// ChangePassword allows a user to change their password
func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := utils.ComparePasswords(user.Password, oldPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password against policy
	if s.passwordPolicyService != nil {
		if err := s.passwordPolicyService.ValidateNewPassword(userID, newPassword, user.Password); err != nil {
			return err
		}
	}

	// Hash new password
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user
	user.Password = hashed
	now := time.Now()
	user.PasswordChangedAt = &now
	user.ChangePassword = false // Clear force change flag

	// Set new expiry
	if s.passwordPolicyService != nil {
		user.PasswordExpiresAt = s.passwordPolicyService.CalculatePasswordExpiry()
	}

	if err := s.repo.Update(user); err != nil {
		return err
	}

	// Record in password history
	if s.passwordPolicyService != nil {
		_ = s.passwordPolicyService.RecordPasswordChange(userID, hashed)
	}

	return nil
}

// ResetPassword allows admin to reset a user's password
func (s *UserService) ResetPassword(userID uuid.UUID, newPassword string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Validate new password against policy
	if s.passwordPolicyService != nil {
		if err := s.passwordPolicyService.ValidateNewPassword(userID, newPassword, ""); err != nil {
			return err
		}
	}

	// Hash new password
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user - force password change on next login
	user.Password = hashed
	now := time.Now()
	user.PasswordChangedAt = &now
	user.ChangePassword = true // Force change on next login
	user.FailedLoginAttempts = 0
	user.IsLocked = false
	user.LockedUntil = nil

	// Set new expiry
	if s.passwordPolicyService != nil {
		user.PasswordExpiresAt = s.passwordPolicyService.CalculatePasswordExpiry()
	}

	if err := s.repo.Update(user); err != nil {
		return err
	}

	// Record in password history
	if s.passwordPolicyService != nil {
		_ = s.passwordPolicyService.RecordPasswordChange(userID, hashed)
	}

	return nil
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

	now := time.Now()
	user := &models.User{
		Email:             email,
		Password:          hashed,
		RoleID:            &role.RoleID,
		IsActive:          true,
		PasswordChangedAt: &now,
	}

	// Set password expiry if policy requires it
	if s.passwordPolicyService != nil {
		user.PasswordExpiresAt = s.passwordPolicyService.CalculatePasswordExpiry()
	}

	if err := s.repo.Create(user); err != nil {
		return err
	}

	// Record password in history
	if s.passwordPolicyService != nil {
		_ = s.passwordPolicyService.RecordPasswordChange(user.UserID, hashed)
	}

	return nil
}
