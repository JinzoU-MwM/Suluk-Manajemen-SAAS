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
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
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
