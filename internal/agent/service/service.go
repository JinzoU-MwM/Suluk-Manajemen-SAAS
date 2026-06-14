package service

import (
	"context"
	"fmt"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/agent/repository"
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
		Type:              req.Type,
		Level:             1,
	}
	if a.CommissionRate <= 0 {
		a.CommissionRate = 5.0
	}
	if a.Type == "" {
		a.Type = "agent"
	}
	if req.ParentID != "" {
		parent, err := s.repo.GetAgent(ctx, req.ParentID, orgID)
		if err != nil || parent == nil {
			return nil, fmt.Errorf("parent agent not found")
		}
		pid := req.ParentID
		a.ParentID = &pid
		a.Level = parent.Level + 1
	}
	if err := s.repo.CreateAgent(ctx, a); err != nil {
		return nil, err
	}
	// Re-read to hydrate parent name + aggregate fields.
	return s.repo.GetAgent(ctx, a.ID, orgID)
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
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			updates["parent_id"] = nil
			updates["level"] = 1
		} else {
			if *req.ParentID == id {
				return nil, fmt.Errorf("agent cannot be its own parent")
			}
			// Cycle guard: the proposed parent's ancestor chain (which includes
			// the proposed parent itself) must not contain this agent.
			ancestors, err := s.repo.AncestorIDs(ctx, *req.ParentID, orgID)
			if err != nil {
				return nil, err
			}
			for _, anc := range ancestors {
				if anc == id {
					return nil, fmt.Errorf("cannot set parent: would create a cycle in the hierarchy")
				}
			}
			parent, err := s.repo.GetAgent(ctx, *req.ParentID, orgID)
			if err != nil || parent == nil {
				return nil, fmt.Errorf("parent agent not found")
			}
			updates["parent_id"] = *req.ParentID
			updates["level"] = parent.Level + 1
		}
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
	c.TierLevel = 1 // the seller's own commission
	// commission.accrued → Dr Beban Komisi / Cr Hutang Komisi. Agent name is
	// best-effort.
	agentName := ""
	if a, aerr := s.repo.GetAgent(ctx, req.AgentID, orgID); aerr == nil && a != nil {
		agentName = a.Name
	}
	sellerPayload := commissionPayload(c.CommissionAmount, agentName, 1, "")
	// Berjenjang: build override commissions for the seller's upline, then write
	// the seller + all tiers (and one event each) atomically.
	tiers := s.buildUplineTiers(ctx, orgID, c, agentName)
	if err := s.repo.CreateCommissionCascadeTx(ctx, c, sellerPayload, tiers); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *AgentService) ListCommissions(ctx context.Context, orgID, agentID, status, tierLevel string, page, limit int) (*model.CommissionListResponse, error) {
	comms, total, err := s.repo.ListCommissions(ctx, orgID, agentID, status, tierLevel, page, limit)
	if err != nil {
		return nil, err
	}
	return &model.CommissionListResponse{Commissions: comms, Total: total}, nil
}

// GetDownline returns the agent's descendant tree (flat, depth-ordered).
func (s *AgentService) GetDownline(ctx context.Context, agentID, orgID string) ([]model.DownlineNode, error) {
	return s.repo.Downline(ctx, agentID, orgID)
}

// GetUpline returns the agent's ancestor chain (nearest-first).
func (s *AgentService) GetUpline(ctx context.Context, agentID, orgID string) ([]model.DownlineNode, error) {
	return s.repo.Upline(ctx, agentID, orgID)
}

// GetTiers returns the org's commission tier rates (defaults if unconfigured).
func (s *AgentService) GetTiers(ctx context.Context, orgID string) ([]model.CommissionTier, error) {
	tiers, err := s.repo.ListTiers(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(tiers) == 0 {
		return defaultTierRates, nil
	}
	return tiers, nil
}

// SetTiers replaces the org's commission tier configuration.
func (s *AgentService) SetTiers(ctx context.Context, orgID string, tiers []model.CommissionTier) error {
	for _, t := range tiers {
		if t.Level < 2 {
			return fmt.Errorf("tier level must be >= 2 (tier 1 is the seller)")
		}
		if t.RatePct < 0 || t.RatePct > 100 {
			return fmt.Errorf("rate_pct must be between 0 and 100")
		}
	}
	return s.repo.UpsertTiers(ctx, orgID, tiers)
}

func (s *AgentService) PayCommission(ctx context.Context, id, orgID string) error {
	return s.repo.PayCommission(ctx, id, orgID)
}

func (s *AgentService) GetAgentCommissions(ctx context.Context, agentID, orgID string) ([]model.AgentCommission, error) {
	return s.repo.GetAgentCommissions(ctx, agentID, orgID)
}
