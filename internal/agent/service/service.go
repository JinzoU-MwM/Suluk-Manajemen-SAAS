package service

import (
	"context"
	"encoding/json"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/agent/repository"
	"github.com/jamaah-in/v2/internal/shared/events"
)

type AgentService struct {
	repo *repository.AgentRepo
}

func NewAgentService(repo *repository.AgentRepo) *AgentService {
	return &AgentService{repo: repo}
}

func (s *AgentService) CreateAgent(ctx context.Context, orgID string, req model.CreateAgentRequest) (*model.Agent, error) {
	a := &model.Agent{
		OrgID:             orgID,
		Name:              req.Name,
		Phone:             req.Phone,
		Email:             req.Email,
		Address:           req.Address,
		CommissionRate:    req.CommissionRate,
		BankName:          req.BankName,
		BankAccountNumber: req.BankAccountNumber,
		BankAccountName:   req.BankAccountName,
		Notes:             req.Notes,
	}
	if a.CommissionRate <= 0 {
		a.CommissionRate = 5.0
	}
	if err := s.repo.CreateAgent(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AgentService) ListAgents(ctx context.Context, orgID, search string, page, limit int) (*model.AgentListResponse, error) {
	agents, total, err := s.repo.ListAgents(ctx, orgID, search, page, limit)
	if err != nil {
		return nil, err
	}
	return &model.AgentListResponse{Agents: agents, Total: total}, nil
}

func (s *AgentService) GetAgent(ctx context.Context, id, orgID string) (*model.Agent, error) {
	return s.repo.GetAgent(ctx, id, orgID)
}

func (s *AgentService) UpdateAgent(ctx context.Context, id, orgID string, req model.UpdateAgentRequest) (*model.Agent, error) {
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.CommissionRate != nil {
		updates["commission_rate"] = *req.CommissionRate
	}
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.BankAccountNumber != nil {
		updates["bank_account_number"] = *req.BankAccountNumber
	}
	if req.BankAccountName != nil {
		updates["bank_account_name"] = *req.BankAccountName
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) == 0 {
		return s.repo.GetAgent(ctx, id, orgID)
	}
	if err := s.repo.UpdateAgent(ctx, id, orgID, updates); err != nil {
		return nil, err
	}
	return s.repo.GetAgent(ctx, id, orgID)
}

func (s *AgentService) CreateCommission(ctx context.Context, orgID string, req model.CreateCommissionRequest) (*model.AgentCommission, error) {
	c := &model.AgentCommission{
		OrgID:            orgID,
		AgentID:          req.AgentID,
		CommissionAmount: req.CommissionAmount,
		CommissionRate:   req.CommissionRate,
		JamaahName:       req.JamaahName,
		PackageName:      req.PackageName,
		Notes:            req.Notes,
	}
	if req.JamaahID != "" {
		c.JamaahID = &req.JamaahID
	}
	if req.InvoiceID != "" {
		c.InvoiceID = &req.InvoiceID
	}
	if req.PackageID != "" {
		c.PackageID = &req.PackageID
	}
	if c.CommissionRate <= 0 {
		c.CommissionRate = 5.0
	}
	// commission.accrued → Dr Beban Komisi / Cr Hutang Komisi. Agent name is
	// best-effort.
	agentName := ""
	if a, aerr := s.repo.GetAgent(ctx, req.AgentID, orgID); aerr == nil && a != nil {
		agentName = a.Name
	}
	payload, _ := json.Marshal(map[string]any{"amount": c.CommissionAmount, "agent_name": agentName})
	if err := s.repo.CreateCommissionTx(ctx, c, events.EventCommissionAccrued, payload); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *AgentService) ListCommissions(ctx context.Context, orgID, agentID, status string, page, limit int) (*model.CommissionListResponse, error) {
	comms, total, err := s.repo.ListCommissions(ctx, orgID, agentID, status, page, limit)
	if err != nil {
		return nil, err
	}
	return &model.CommissionListResponse{Commissions: comms, Total: total}, nil
}

func (s *AgentService) PayCommission(ctx context.Context, id, orgID string) error {
	return s.repo.PayCommission(ctx, id, orgID)
}

func (s *AgentService) GetAgentCommissions(ctx context.Context, agentID, orgID string) ([]model.AgentCommission, error) {
	return s.repo.GetAgentCommissions(ctx, agentID, orgID)
}
