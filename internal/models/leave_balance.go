package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaveBalance struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	LeaveTypeID   uuid.UUID `json:"leave_type_id"`
	Year          int       `json:"year"`
	TotalEntitled int       `json:"total_entitled"`
	Used          int       `json:"used"`
	Pending       int       `json:"pending"`
	CarriedForward int      `json:"carried_forward"`
	Adjustment    int       `json:"adjustment"`
	// Balance is computed: total_entitled + carried_forward + adjustment - used - pending
	Balance    int       `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations (populated on demand)
	LeaveType *LeaveType `json:"leave_type,omitempty"`
}
