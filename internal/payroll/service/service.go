package service

import (
	"context"

	"github.com/jamaah-in/v2/internal/payroll/model"
	"github.com/jamaah-in/v2/internal/payroll/repository"
)

type PayrollService struct {
	repo *repository.PayrollRepo
}

func NewPayrollService(repo *repository.PayrollRepo) *PayrollService {
	return &PayrollService{repo: repo}
}

func (s *PayrollService) CreateEmployee(ctx context.Context, orgID string, req model.CreateEmployeeRequest) (*model.Employee, error) {
	e := &model.Employee{
		OrgID:      orgID,
		Name:       req.Name,
		Position:   req.Position,
		Type:       req.Type,
		BaseSalary: req.BaseSalary,
		Allowance:  req.Allowance,
		BpjsTk:     req.BpjsTk,
		BpjsKes:    req.BpjsKes,
		Pph21Rate:  req.Pph21Rate,
		Phone:      req.Phone,
		Email:      req.Email,
		IsActive:   true,
	}
	if e.Type == "" {
		e.Type = "tetap"
	}
	if err := s.repo.CreateEmployee(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *PayrollService) ListEmployees(ctx context.Context, orgID string) ([]model.Employee, error) {
	return s.repo.ListEmployees(ctx, orgID)
}

func (s *PayrollService) GetEmployee(ctx context.Context, id, orgID string) (*model.Employee, error) {
	return s.repo.GetEmployee(ctx, id, orgID)
}

func (s *PayrollService) UpdateEmployee(ctx context.Context, id, orgID string, req model.UpdateEmployeeRequest) (*model.Employee, error) {
	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Position != nil { updates["position"] = *req.Position }
	if req.Type != nil { updates["type"] = *req.Type }
	if req.BaseSalary != nil { updates["base_salary"] = *req.BaseSalary }
	if req.Allowance != nil { updates["allowance"] = *req.Allowance }
	if req.BpjsTk != nil { updates["bpjs_tk"] = *req.BpjsTk }
	if req.BpjsKes != nil { updates["bpjs_kes"] = *req.BpjsKes }
	if req.Pph21Rate != nil { updates["pph21_rate"] = *req.Pph21Rate }
	if req.Phone != nil { updates["phone"] = *req.Phone }
	if req.Email != nil { updates["email"] = *req.Email }
	if req.IsActive != nil { updates["is_active"] = *req.IsActive }
	if len(updates) == 0 {
		return s.repo.GetEmployee(ctx, id, orgID)
	}
	if err := s.repo.UpdateEmployee(ctx, id, orgID, updates); err != nil {
		return nil, err
	}
	return s.repo.GetEmployee(ctx, id, orgID)
}

func (s *PayrollService) CreateSalarySlip(ctx context.Context, orgID string, req model.CreateSalarySlipRequest) (*model.SalarySlip, error) {
	emp, err := s.repo.GetEmployee(ctx, req.EmployeeID, orgID)
	if err != nil {
		return nil, err
	}

	gross := emp.BaseSalary + emp.Allowance
	pph21 := int64(float64(gross) * emp.Pph21Rate / 100)
	bpjs := emp.BpjsTk + emp.BpjsKes
	net := gross - pph21 - bpjs

	slip := &model.SalarySlip{
		OrgID:      orgID,
		EmployeeID: req.EmployeeID,
		Period:     req.Period,
		BaseSalary: emp.BaseSalary,
		Allowance:  emp.Allowance,
		Pph21Amount: pph21,
		BpjsAmount: bpjs,
		NetSalary:  net,
		Notes:      req.Notes,
	}
	if req.PackageID != "" {
		slip.PackageID = &req.PackageID
	}
	if err := s.repo.CreateSalarySlip(ctx, slip); err != nil {
		return nil, err
	}
	return slip, nil
}

func (s *PayrollService) ListSalarySlips(ctx context.Context, orgID, period string) ([]model.SalarySlip, error) {
	return s.repo.ListSalarySlips(ctx, orgID, period)
}

func (s *PayrollService) FinalizeSlip(ctx context.Context, id, orgID string) error {
	return s.repo.UpdateSalarySlipStatus(ctx, id, orgID, "final")
}

func (s *PayrollService) CreateAdvance(ctx context.Context, orgID string, req model.CreateAdvanceRequest) (*model.Advance, error) {
	a := &model.Advance{
		OrgID:      orgID,
		EmployeeID: req.EmployeeID,
		Amount:     req.Amount,
		Remaining:  req.Amount,
		Reason:     req.Reason,
		Status:     "active",
	}
	if err := s.repo.CreateAdvance(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PayrollService) ListAdvances(ctx context.Context, orgID string) ([]model.Advance, error) {
	return s.repo.ListAdvances(ctx, orgID)
}

func (s *PayrollService) RepayAdvance(ctx context.Context, id, orgID string, req model.RepayAdvanceRequest) error {
	var slipID *string
	if req.SalarySlipID != "" {
		slipID = &req.SalarySlipID
	}
	return s.repo.RepayAdvance(ctx, id, req.Amount, slipID)
}

func (s *PayrollService) GetPayrollSummary(ctx context.Context, orgID string) (*model.PayrollSummary, error) {
	return s.repo.GetPayrollSummary(ctx, orgID)
}
