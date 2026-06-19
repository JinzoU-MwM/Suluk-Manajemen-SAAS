package model

import "time"

type MemberEquipment struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"org_id"`
	PackageID           string     `json:"package_id"`
	MemberID            string     `json:"member_id"`
	Nama                string     `json:"nama"`
	Gender              string     `json:"gender"`
	BajuSize            string     `json:"baju_size"`
	FamilyID            string     `json:"family_id"`
	IsEquipmentReceived bool       `json:"is_equipment_received"`
	ReceivedItems       []string   `json:"received_items"`
	ReceivedAt          *time.Time `json:"received_at"`
	HandoverToken       string     `json:"handover_token"`
	IsLuggageChecked    bool       `json:"is_luggage_checked"`
	LuggageCheckedAt    *time.Time `json:"luggage_checked_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ScanRequest records a QR handover scan for a member (looked up by token).
type ScanRequest struct {
	Token      string   `json:"token"`
	Checkpoint string   `json:"checkpoint"` // equipment|luggage
	Items      []string `json:"items,omitempty"`
}

// CheckpointMember is a member's handover progress for the package view.
type CheckpointMember struct {
	MemberID            string `json:"member_id"`
	Nama                string `json:"nama"`
	HandoverToken       string `json:"handover_token"`
	IsEquipmentReceived bool   `json:"is_equipment_received"`
	IsLuggageChecked    bool   `json:"is_luggage_checked"`
}

type InventoryItem struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Unit      string    `json:"unit"`
	Stock     int       `json:"stock"`
	MinStock  int       `json:"min_stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ForecastResponse struct {
	TotalMembers  int            `json:"total_members"`
	Requirements  map[string]int `json:"requirements"`
	SizeBreakdown map[string]int `json:"size_breakdown"`
	Details       []MemberDetail `json:"details"`
}

type MemberDetail struct {
	MemberID            string `json:"member_id"`
	Nama                string `json:"nama"`
	Gender              string `json:"gender"`
	BajuSize            string `json:"baju_size"`
	FamilyID            string `json:"family_id"`
	IsEquipmentReceived bool   `json:"is_equipment_received"`
}

type FulfillmentResponse struct {
	ReceivedCount int           `json:"received_count"`
	PendingCount  int           `json:"pending_count"`
	Received      []MemberBrief `json:"received"`
	Pending       []MemberBrief `json:"pending"`
}

type MemberBrief struct {
	ID                  string `json:"id"`
	Nama                string `json:"nama"`
	IsEquipmentReceived bool   `json:"is_equipment_received"`
}

type SyncMembersRequest struct {
	PackageID string       `json:"package_id"`
	Members   []SyncMember `json:"members"`
}

type SyncMember struct {
	MemberID string `json:"member_id"`
	Nama     string `json:"nama"`
	Gender   string `json:"gender"`
	BajuSize string `json:"baju_size"`
	FamilyID string `json:"family_id"`
}

type MarkReceivedRequest struct {
	MemberIDs     []string `json:"member_ids"`
	ItemsReceived []string `json:"items_received"`
}

type UpdateOperationalRequest struct {
	BajuSize string `json:"baju_size"`
	FamilyID string `json:"family_id"`
}

// --- Stock monitoring (Phase 6) ---

type StockItem struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Unit      string    `json:"unit"`
	Stock     int       `json:"stock"`
	MinStock  int       `json:"min_stock"`
	InKit     bool      `json:"in_kit"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Delta     int       `json:"delta"`
	Reason    string    `json:"reason"`
	Note      string    `json:"note"`
	GroupID   *string   `json:"group_id,omitempty"`
	PackageID *string   `json:"package_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type PackageKitItem struct {
	ItemID       string `json:"item_id"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	QtyPerJamaah int    `json:"qty_per_jamaah"`
}

// Deduction is one computed (item, quantity) to subtract — pure logic output.
type Deduction struct {
	ItemID string
	Qty    int
}

// DepartedPayload is the group.departed event payload.
type DepartedPayload struct {
	GroupID     string `json:"group_id"`
	PackageID   string `json:"package_id"`
	MemberCount int    `json:"member_count"`
	Status      string `json:"status"`
}

type CreateItemRequest struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Unit         string `json:"unit"`
	MinStock     int    `json:"min_stock"`
	InitialStock int    `json:"initial_stock"`
}

type UpdateItemRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
	MinStock int    `json:"min_stock"`
}

type RestockRequest struct {
	Qty  int    `json:"qty"`
	Note string `json:"note"`
}

type AdjustRequest struct {
	Delta int    `json:"delta"`
	Note  string `json:"note"`
}

type KitLine struct {
	ItemID       string `json:"item_id"`
	QtyPerJamaah int    `json:"qty_per_jamaah"`
}

type SetKitRequest struct {
	Items []KitLine `json:"items"`
}
