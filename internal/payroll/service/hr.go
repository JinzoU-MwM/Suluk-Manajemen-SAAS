package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jamaah-in/v2/internal/payroll/model"
)

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func (s *PayrollService) RecordAttendance(ctx context.Context, orgID string, req model.RecordAttendanceRequest) (*model.Attendance, error) {
	if req.EmployeeID == "" || req.Date == "" {
		return nil, fmt.Errorf("employee_id and date are required")
	}
	if _, err := parseDate(req.Date); err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}
	status := req.Status
	if status == "" {
		status = "hadir"
	}
	a := &model.Attendance{OrgID: orgID, EmployeeID: req.EmployeeID, Date: req.Date, Status: status, Notes: req.Notes}
	if err := s.repo.RecordAttendance(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PayrollService) ListAttendance(ctx context.Context, orgID, employeeID, period string) ([]model.Attendance, error) {
	return s.repo.ListAttendance(ctx, orgID, employeeID, period)
}

func (s *PayrollService) AttendanceSummary(ctx context.Context, orgID, employeeID, period string) (*model.AttendanceSummary, error) {
	return s.repo.AttendanceSummary(ctx, orgID, employeeID, period)
}

func (s *PayrollService) CreateLeave(ctx context.Context, orgID string, req model.CreateLeaveRequest) (*model.LeaveRequest, error) {
	if req.EmployeeID == "" {
		return nil, fmt.Errorf("employee_id is required")
	}
	start, err := parseDate(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("start_date: %w", err)
	}
	end, err := parseDate(req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("end_date: %w", err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date sebelum start_date")
	}
	days := int(end.Sub(start).Hours()/24) + 1 // inclusive
	typ := req.Type
	if typ == "" {
		typ = "tahunan"
	}
	l := &model.LeaveRequest{
		OrgID: orgID, EmployeeID: req.EmployeeID, Type: typ,
		StartDate: req.StartDate, EndDate: req.EndDate, Days: days, Reason: req.Reason,
	}
	if err := s.repo.CreateLeave(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *PayrollService) ListLeave(ctx context.Context, orgID, status string) ([]model.LeaveRequest, error) {
	return s.repo.ListLeave(ctx, orgID, status)
}

func (s *PayrollService) DecideLeave(ctx context.Context, id, orgID, decidedBy string, req model.DecideLeaveRequest) error {
	if req.Status != "approved" && req.Status != "rejected" {
		return fmt.Errorf("status harus approved atau rejected")
	}
	return s.repo.DecideLeave(ctx, id, orgID, req.Status, decidedBy)
}
