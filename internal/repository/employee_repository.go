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

type EmployeeRepository struct {
	db *sql.DB
}

func NewEmployeeRepository() *EmployeeRepository {
	return &EmployeeRepository{db: database.DB}
}

const employeeSelectCols = `
	e.id, e.user_id, e.employee_number, e.first_name, e.last_name, e.email, e.personal_email,
	e.phone, e.date_of_birth, e.gender, e.national_id, e.marital_status, e.address, e.city, e.state,
	e.country, e.department_id, e.position_id, e.manager_id, e.hire_date, e.probation_end_date,
	e.employment_type, e.employment_status, e.termination_date, e.termination_reason,
	e.profile_photo_url, e.created_at, e.updated_at, e.deleted_at`

func (r *EmployeeRepository) Create(emp *models.Employee) error {
	emp.ID = uuid.New()
	now := time.Now()
	emp.CreatedAt = now
	emp.UpdatedAt = now
	_, err := r.db.Exec(`
		INSERT INTO employees (id, user_id, employee_number, first_name, last_name, email, personal_email,
		phone, date_of_birth, gender, national_id, marital_status, address, city, state, country,
		department_id, position_id, manager_id, hire_date, probation_end_date, employment_type,
		employment_status, termination_date, termination_reason, profile_photo_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
		emp.ID, emp.UserID, emp.EmployeeNumber, emp.FirstName, emp.LastName, emp.Email, emp.PersonalEmail,
		emp.Phone, emp.DateOfBirth, emp.Gender, emp.NationalID, emp.MaritalStatus, emp.Address, emp.City,
		emp.State, emp.Country, emp.DepartmentID, emp.PositionID, emp.ManagerID, emp.HireDate,
		emp.ProbationEndDate, emp.EmploymentType, emp.EmploymentStatus, emp.TerminationDate,
		emp.TerminationReason, emp.ProfilePhotoURL, emp.CreatedAt, emp.UpdatedAt,
	)
	return err
}

func (r *EmployeeRepository) GetByID(id uuid.UUID) (*models.Employee, error) {
	query := fmt.Sprintf(`SELECT %s FROM employees e WHERE e.id=$1 AND e.deleted_at IS NULL`, employeeSelectCols)
	row := r.db.QueryRow(query, id)
	return r.scanEmployee(row)
}

func (r *EmployeeRepository) GetByEmployeeNumber(number string) (*models.Employee, error) {
	query := fmt.Sprintf(`SELECT %s FROM employees e WHERE e.employee_number=$1 AND e.deleted_at IS NULL`, employeeSelectCols)
	row := r.db.QueryRow(query, number)
	return r.scanEmployee(row)
}

func (r *EmployeeRepository) List(filter interfaces.EmployeeFilter, page, pageSize int) ([]models.Employee, int, error) {
	args := []interface{}{}
	where := []string{}
	i := 1

	if !filter.IncludeDeleted {
		where = append(where, "e.deleted_at IS NULL")
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf(
			"(e.first_name ILIKE $%d OR e.last_name ILIKE $%d OR e.email ILIKE $%d OR e.employee_number ILIKE $%d)",
			i, i, i, i))
		args = append(args, "%"+filter.Search+"%")
		i++
	}
	if filter.DepartmentID != nil {
		where = append(where, fmt.Sprintf("e.department_id=$%d", i))
		args = append(args, *filter.DepartmentID)
		i++
	}
	if filter.PositionID != nil {
		where = append(where, fmt.Sprintf("e.position_id=$%d", i))
		args = append(args, *filter.PositionID)
		i++
	}
	if filter.EmploymentStatus != "" {
		where = append(where, fmt.Sprintf("e.employment_status=$%d", i))
		args = append(args, filter.EmploymentStatus)
		i++
	}
	if filter.EmploymentType != "" {
		where = append(where, fmt.Sprintf("e.employment_type=$%d", i))
		args = append(args, filter.EmploymentType)
		i++
	}

	whereStr := "1=1"
	if len(where) > 0 {
		whereStr = strings.Join(where, " AND ")
	}

	var total int
	err := r.db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM employees e WHERE %s`, whereStr), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT %s FROM employees e WHERE %s ORDER BY e.last_name, e.first_name LIMIT $%d OFFSET $%d`,
		employeeSelectCols, whereStr, i, i+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var emps []models.Employee
	for rows.Next() {
		emp, err := r.scanEmployee(rows)
		if err != nil {
			return nil, 0, err
		}
		emps = append(emps, *emp)
	}
	return emps, total, rows.Err()
}

func (r *EmployeeRepository) Update(emp *models.Employee) error {
	emp.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE employees SET first_name=$1, last_name=$2, email=$3, personal_email=$4, phone=$5,
		date_of_birth=$6, gender=$7, national_id=$8, marital_status=$9, address=$10, city=$11, state=$12,
		country=$13, department_id=$14, position_id=$15, manager_id=$16, probation_end_date=$17,
		employment_type=$18, employment_status=$19, termination_date=$20, termination_reason=$21,
		profile_photo_url=$22, updated_at=$23 WHERE id=$24 AND deleted_at IS NULL`,
		emp.FirstName, emp.LastName, emp.Email, emp.PersonalEmail, emp.Phone,
		emp.DateOfBirth, emp.Gender, emp.NationalID, emp.MaritalStatus, emp.Address, emp.City,
		emp.State, emp.Country, emp.DepartmentID, emp.PositionID, emp.ManagerID, emp.ProbationEndDate,
		emp.EmploymentType, emp.EmploymentStatus, emp.TerminationDate, emp.TerminationReason,
		emp.ProfilePhotoURL, emp.UpdatedAt, emp.ID,
	)
	return err
}

func (r *EmployeeRepository) SoftDelete(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE employees SET deleted_at=$1, updated_at=$1 WHERE id=$2 AND deleted_at IS NULL`, time.Now(), id)
	return err
}

func (r *EmployeeRepository) GetDirectReports(managerID uuid.UUID) ([]models.Employee, error) {
	query := fmt.Sprintf(`SELECT %s FROM employees e WHERE e.manager_id=$1 AND e.deleted_at IS NULL ORDER BY e.last_name`, employeeSelectCols)
	rows, err := r.db.Query(query, managerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []models.Employee
	for rows.Next() {
		emp, err := r.scanEmployee(rows)
		if err != nil {
			return nil, err
		}
		emps = append(emps, *emp)
	}
	return emps, rows.Err()
}

func (r *EmployeeRepository) CountByDatePrefix(prefix string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM employees WHERE employee_number LIKE $1`, "EMP-"+prefix+"%").Scan(&count)
	return count, err
}

func (r *EmployeeRepository) EmailActiveExists(email string, excludeID *uuid.UUID) (bool, error) {
	var count int
	var err error
	if excludeID != nil {
		err = r.db.QueryRow(`SELECT COUNT(1) FROM employees WHERE email=$1 AND id!=$2 AND deleted_at IS NULL`, email, excludeID).Scan(&count)
	} else {
		err = r.db.QueryRow(`SELECT COUNT(1) FROM employees WHERE email=$1 AND deleted_at IS NULL`, email).Scan(&count)
	}
	return count > 0, err
}

func (r *EmployeeRepository) scanEmployee(row rowScanner) (*models.Employee, error) {
	var e models.Employee
	var userID, managerID sql.NullString
	var dob, probEnd, termDate sql.NullTime

	err := row.Scan(
		&e.ID, &userID, &e.EmployeeNumber, &e.FirstName, &e.LastName, &e.Email, &e.PersonalEmail,
		&e.Phone, &dob, &e.Gender, &e.NationalID, &e.MaritalStatus, &e.Address, &e.City, &e.State,
		&e.Country, &e.DepartmentID, &e.PositionID, &managerID, &e.HireDate, &probEnd,
		&e.EmploymentType, &e.EmploymentStatus, &termDate, &e.TerminationReason,
		&e.ProfilePhotoURL, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		p, _ := uuid.Parse(userID.String)
		e.UserID = &p
	}
	if managerID.Valid {
		m, _ := uuid.Parse(managerID.String)
		e.ManagerID = &m
	}
	if dob.Valid {
		e.DateOfBirth = &dob.Time
	}
	if probEnd.Valid {
		e.ProbationEndDate = &probEnd.Time
	}
	if termDate.Valid {
		e.TerminationDate = &termDate.Time
	}

	return &e, nil
}
