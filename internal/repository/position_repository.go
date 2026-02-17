package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"hr-system/internal/database"
	"hr-system/internal/interfaces"
	"hr-system/internal/models"

	"github.com/google/uuid"
)

type PositionRepository struct {
	db *sql.DB
}

func NewPositionRepository() *PositionRepository {
	return &PositionRepository{db: database.DB}
}

func (r *PositionRepository) Create(pos *models.Position) error {
	pos.ID = uuid.New()
	now := time.Now()
	pos.CreatedAt = now
	pos.UpdatedAt = now
	_, err := r.db.Exec(`
		INSERT INTO positions (id, title, code, department_id, grade_level, min_salary, max_salary, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		pos.ID, pos.Title, pos.Code, pos.DepartmentID, pos.GradeLevel,
		pos.MinSalary, pos.MaxSalary, pos.Description, pos.IsActive, pos.CreatedAt, pos.UpdatedAt,
	)
	return err
}

func (r *PositionRepository) GetByID(id uuid.UUID) (*models.Position, error) {
	var p models.Position
	err := r.db.QueryRow(`
		SELECT id, title, code, department_id, grade_level, min_salary, max_salary, description, is_active, created_at, updated_at, deleted_at
		FROM positions WHERE id=$1 AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.Title, &p.Code, &p.DepartmentID, &p.GradeLevel,
		&p.MinSalary, &p.MaxSalary, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PositionRepository) List(filter interfaces.PositionFilter, page, pageSize int) ([]models.Position, int, error) {
	args := []interface{}{}
	where := []string{"deleted_at IS NULL"}
	i := 1

	if filter.DepartmentID != nil {
		where = append(where, fmt.Sprintf("department_id=$%d", i))
		args = append(args, *filter.DepartmentID)
		i++
	}
	if filter.GradeLevel != "" {
		where = append(where, fmt.Sprintf("grade_level=$%d", i))
		args = append(args, filter.GradeLevel)
		i++
	}
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", i))
		args = append(args, *filter.IsActive)
		i++
	}

	whereStr := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM positions WHERE %s`, whereStr), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT id, title, code, department_id, grade_level, min_salary, max_salary, description, is_active, created_at, updated_at, deleted_at
		FROM positions WHERE %s ORDER BY title LIMIT $%d OFFSET $%d`, whereStr, i, i+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		if err := rows.Scan(&p.ID, &p.Title, &p.Code, &p.DepartmentID, &p.GradeLevel,
			&p.MinSalary, &p.MaxSalary, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, 0, err
		}
		positions = append(positions, p)
	}
	return positions, total, rows.Err()
}

func (r *PositionRepository) Update(pos *models.Position) error {
	pos.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE positions SET title=$1, code=$2, department_id=$3, grade_level=$4, min_salary=$5,
		max_salary=$6, description=$7, is_active=$8, updated_at=$9 WHERE id=$10 AND deleted_at IS NULL`,
		pos.Title, pos.Code, pos.DepartmentID, pos.GradeLevel, pos.MinSalary,
		pos.MaxSalary, pos.Description, pos.IsActive, pos.UpdatedAt, pos.ID,
	)
	return err
}

func (r *PositionRepository) SoftDelete(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE positions SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL`, time.Now(), id)
	return err
}

func (r *PositionRepository) CodeExists(code string, excludeID *uuid.UUID) (bool, error) {
	var count int
	if excludeID != nil {
		err := r.db.QueryRow(`SELECT COUNT(1) FROM positions WHERE code=$1 AND id!=$2 AND deleted_at IS NULL`, code, excludeID).Scan(&count)
		return count > 0, err
	}
	err := r.db.QueryRow(`SELECT COUNT(1) FROM positions WHERE code=$1 AND deleted_at IS NULL`, code).Scan(&count)
	return count > 0, err
}

func (r *PositionRepository) ActiveEmployeeCount(id uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM employees WHERE position_id=$1 AND deleted_at IS NULL AND employment_status='active'`, id).Scan(&count)
	return count, err
}
