package models

import (
	"time"

	"github.com/google/uuid"
)

type Position struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Code         string     `json:"code"`
	DepartmentID uuid.UUID  `json:"department_id"`
	RoleID       *uuid.UUID `json:"role_id,omitempty"` // System role assigned to this position
	GradeLevel   string     `json:"grade_level"`
	MinSalary    string     `json:"min_salary"`
	MaxSalary    string     `json:"max_salary"`
	Description  string     `json:"description"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	// Relations (populated on demand)
	Role *Role `json:"role,omitempty"`
}
