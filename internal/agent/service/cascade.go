package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/jamaah-in/v2/internal/agent/model"
	"github.com/jamaah-in/v2/internal/agent/repository"
)

// defaultTierRates are the upline override rates used when an org hasn't
// configured commission_tiers. Tier 1 (the seller) is not listed — they receive
// the base commission as entered.
var defaultTierRates = []model.CommissionTier{
	{Level: 2, RatePct: 2.0},
	{Level: 3, RatePct: 1.0},
}

// saleBase reverses the seller's commission back to the sale value it was
// computed from (amount / rate). Upline tiers apply their own rate to this same
// base. When the rate is unknown the amount itself is used as the base.
func saleBase(amount int64, ratePct float64) int64 {
	if ratePct <= 0 {
		return amount
	}
	return int64(math.Round(float64(amount) * 100.0 / ratePct))
}

// tieredAmount is one upline tier's commission: ratePct of the sale base.
func tieredAmount(base int64, ratePct float64) int64 {
	if base <= 0 || ratePct <= 0 {
		return 0
	}
	return int64(math.Round(float64(base) * ratePct / 100.0))
}

// tierRates returns the org's configured upline rates keyed by tier level,
// falling back to defaults when none are set.
func (s *AgentService) tierRates(ctx context.Context, orgID string) map[int]float64 {
	tiers, err := s.repo.ListTiers(ctx, orgID)
	if err != nil || len(tiers) == 0 {
		tiers = defaultTierRates
	}
	m := make(map[int]float64, len(tiers))
	for _, t := range tiers {
		m[t.Level] = t.RatePct
	}
	return m
}

// buildUplineTiers walks the seller's upline and builds one override commission
// per configured tier level that has a live ancestor. Pure given the repo reads;
// the actual scoring math (saleBase/tieredAmount) is unit-tested directly.
func (s *AgentService) buildUplineTiers(ctx context.Context, orgID string, seller *model.AgentCommission, sellerAgentName string) []repository.TierCommission {
	rates := s.tierRates(ctx, orgID)
	maxLevel := 0
	for lvl := range rates {
		if lvl > maxLevel {
			maxLevel = lvl
		}
	}
	if maxLevel < 2 {
		return nil
	}

	uplines, err := s.repo.UplineAgents(ctx, seller.AgentID, orgID, maxLevel-1)
	if err != nil || len(uplines) == 0 {
		return nil
	}

	base := saleBase(seller.CommissionAmount, seller.CommissionRate)
	out := make([]repository.TierCommission, 0, len(uplines))
	for _, up := range uplines {
		if !up.IsActive {
			continue // deactivated upline agents don't accrue override commissions
		}
		tierLevel := up.Depth + 1 // depth 1 (direct parent) = tier 2
		rate, ok := rates[tierLevel]
		if !ok || rate <= 0 {
			continue
		}
		amt := tieredAmount(base, rate)
		if amt <= 0 {
			continue
		}
		c := &model.AgentCommission{
			OrgID:            orgID,
			AgentID:          up.ID,
			JamaahID:         seller.JamaahID,
			InvoiceID:        seller.InvoiceID,
			PackageID:        seller.PackageID,
			JamaahName:       seller.JamaahName,
			PackageName:      seller.PackageName,
			CommissionAmount: amt,
			CommissionRate:   rate,
			Notes:            fmt.Sprintf("Komisi berjenjang tier %d (override dari %s)", tierLevel, sellerAgentName),
			TierLevel:        tierLevel,
		}
		out = append(out, repository.TierCommission{
			Commission: c,
			Payload:    commissionPayload(amt, up.Name, tierLevel, sellerAgentName),
		})
	}
	return out
}

// commissionPayload is the commission.accrued event body the accounting service
// consumes. tier/source_agent_name are additive (older payloads omit them).
func commissionPayload(amount int64, agentName string, tier int, sourceAgentName string) []byte {
	m := map[string]any{"amount": amount, "agent_name": agentName, "tier": tier}
	if sourceAgentName != "" {
		m["source_agent_name"] = sourceAgentName
	}
	b, _ := json.Marshal(m)
	return b
}
