package repository

import (
	"database/sql"
	"time"

	"hr-system/internal/database"
	"hr-system/internal/models"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.DB}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (user_id, email, password, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	user.UserID = uuid.New()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.db.Exec(query,
		user.UserID, user.Email, user.Password, user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.user_id, u.email, u.password, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.role_id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.role_id
		WHERE u.user_id = $1`

	return r.scanUser(r.db.QueryRow(query, id))
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT u.user_id, u.email, u.password, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.role_id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.role_id
		WHERE u.email = $1`

	return r.scanUser(r.db.QueryRow(query, email))
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT u.user_id, u.email, u.password, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.role_id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.role_id
		ORDER BY u.created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET email = $1, role_id = $2, is_active = $3, updated_at = $4
		WHERE user_id = $5`
	_, err := r.db.Exec(query, user.Email, user.RoleID, user.IsActive, time.Now(), user.UserID)
	return err
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM users WHERE email = $1`, email).Scan(&count)
	return count > 0, err
}

// scanUser works with both *sql.Row and *sql.Rows via a common interface
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *UserRepository) scanUser(row rowScanner) (*models.User, error) {
	var u models.User
	var roleID sql.NullString
	var rRoleID, rName, rDesc sql.NullString

	err := row.Scan(
		&u.UserID, &u.Email, &u.Password, &roleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		&rRoleID, &rName, &rDesc,
	)
	if err != nil {
		return nil, err
	}

	if roleID.Valid {
		parsed, _ := uuid.Parse(roleID.String)
		u.RoleID = &parsed
	}

	if rRoleID.Valid {
		roleUUID, _ := uuid.Parse(rRoleID.String)
		u.Role = &models.Role{
			RoleID:      roleUUID,
			Name:        rName.String,
			Description: rDesc.String,
		}
	}

	return &u, nil
}

// GetUserWithFewestTasksByRole finds a user with the specified role who has the fewest pending tasks
// This enables load balancing when assigning workflow tasks
func (r *UserRepository) GetUserWithFewestTasksByRole(roleName string) (*models.User, error) {
	query := `
		SELECT u.user_id, u.email, u.password, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.role_id, r.name, r.description,
		       COALESCE(COUNT(at.id), 0) as task_count
		FROM users u
		INNER JOIN roles r ON u.role_id = r.role_id
		LEFT JOIN assigned_tasks at ON at.assigned_to = u.user_id AND at.status = 'pending'
		WHERE r.name = $1 AND u.is_active = true
		GROUP BY u.user_id, u.email, u.password, u.role_id, u.is_active, u.created_at, u.updated_at,
		         r.role_id, r.name, r.description
		ORDER BY task_count ASC, u.created_at ASC
		LIMIT 1
	`

	var u models.User
	var roleID sql.NullString
	var rRoleID, rName, rDesc sql.NullString
	var taskCount int

	err := r.db.QueryRow(query, roleName).Scan(
		&u.UserID, &u.Email, &u.Password, &roleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		&rRoleID, &rName, &rDesc, &taskCount,
	)
	if err != nil {
		return nil, err
	}

	if roleID.Valid {
		parsed, _ := uuid.Parse(roleID.String)
		u.RoleID = &parsed
	}

	if rRoleID.Valid {
		roleUUID, _ := uuid.Parse(rRoleID.String)
		u.Role = &models.Role{
			RoleID:      roleUUID,
			Name:        rName.String,
			Description: rDesc.String,
		}
	}

	return &u, nil
}
