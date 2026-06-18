package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// normalizeGender maps assorted gender spellings to a rooming bucket label.
func normalizeGender(g string) string {
	switch strings.ToUpper(strings.TrimSpace(g)) {
	case "L", "LAKI-LAKI", "PRIA", "M", "MALE":
		return "Ikhwan"
	case "P", "PEREMPUAN", "WANITA", "F", "FEMALE":
		return "Akhwat"
	default:
		return "Lainnya"
	}
}

func roomTypeForCapacity(capacity int) string {
	switch capacity {
	case 1:
		return "single"
	case 2:
		return "double"
	case 3:
		return "triple"
	case 4:
		return "quad"
	default:
		return "custom"
	}
}

func (s *JamaahService) ListRooms(ctx context.Context, orgID uuid.UUID, groupID *uuid.UUID) ([]model.Room, error) {
	rooms, err := s.repo.ListRooms(ctx, orgID, groupID)
	if err != nil {
		return nil, err
	}
	if rooms == nil {
		return []model.Room{}, nil
	}
	return rooms, nil
}

func (s *JamaahService) CreateRoom(ctx context.Context, orgID uuid.UUID, groupID *uuid.UUID, req model.CreateRoomRequest) (*model.Room, error) {
	if req.RoomNumber == "" {
		return nil, fmt.Errorf("room_number is required")
	}
	room := &model.Room{
		ID:         uuid.New().String(),
		OrgID:      orgID.String(),
		RoomNumber: req.RoomNumber,
		GenderType: req.GenderType,
		RoomType:   req.RoomType,
		Capacity:   req.Capacity,
		IsActive:   true,
	}
	if groupID != nil {
		gs := groupID.String()
		room.GroupID = &gs
	}
	if room.GenderType == "" {
		room.GenderType = "mixed"
	}
	if room.RoomType == "" {
		room.RoomType = "double"
	}
	if room.Capacity < 1 {
		room.Capacity = 2
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *JamaahService) DeleteRoom(ctx context.Context, roomID, orgID uuid.UUID) error {
	return s.repo.DeleteRoom(ctx, roomID, orgID)
}

// AutoRooming regenerates rooms for a group: members are split by gender
// (never mixed — required for Umrah/Haji) and packed into rooms of `capacity`.
// Members keep their group order within a gender, so a family added together
// tends to land in the same room. capacity falls back to 4 (quad) when invalid.
func (s *JamaahService) AutoRooming(ctx context.Context, orgID, groupID uuid.UUID, capacity int) ([]model.Room, error) {
	if capacity < 1 || capacity > 6 {
		capacity = 4
	}
	_ = s.repo.DeleteRoomsByGroup(ctx, groupID, orgID)

	members, err := s.repo.ListGroupMembersWithGender(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Bucket members by normalized gender, preserving first-seen order.
	buckets := map[string][]model.RoomCandidate{}
	var order []string
	for _, m := range members {
		g := normalizeGender(m.Gender)
		if _, ok := buckets[g]; !ok {
			order = append(order, g)
		}
		buckets[g] = append(buckets[g], m)
	}

	rooms := []model.Room{}
	roomNum := 1
	for _, g := range order {
		list := buckets[g]
		for i := 0; i < len(list); i += capacity {
			gs := groupID.String()
			room := &model.Room{
				ID:         uuid.New().String(),
				OrgID:      orgID.String(),
				GroupID:    &gs,
				RoomNumber: fmt.Sprintf("%d", roomNum),
				GenderType: g,
				RoomType:   roomTypeForCapacity(capacity),
				Capacity:   capacity,
				IsActive:   true,
			}
			if err := s.repo.CreateRoom(ctx, room); err == nil {
				rooms = append(rooms, *room)
				roomID, _ := uuid.Parse(room.ID)
				for j := i; j < i+capacity && j < len(list); j++ {
					s.repo.AssignMemberToRoom(ctx, orgID, roomID, list[j].MemberID.String())
				}
			}
			roomNum++
		}
	}
	return rooms, nil
}

func (s *JamaahService) ClearAutoRooming(ctx context.Context, orgID, groupID uuid.UUID) error {
	return s.repo.DeleteRoomsByGroup(ctx, groupID, orgID)
}

func (s *JamaahService) AssignMemberToRoom(ctx context.Context, orgID, roomID uuid.UUID, memberID string) error {
	return s.repo.AssignMemberToRoom(ctx, orgID, roomID, memberID)
}

func (s *JamaahService) UnassignMember(ctx context.Context, orgID uuid.UUID, memberID string) error {
	return s.repo.UnassignMember(ctx, orgID, memberID)
}

func (s *JamaahService) ShareGroup(ctx context.Context, orgID, groupID uuid.UUID, pin string, expiresInDays int) (*model.SharedManifest, error) {
	_ = s.repo.RevokeSharedManifest(ctx, groupID, orgID)

	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	gs := groupID.String()
	expiresAt := time.Now().AddDate(0, 0, expiresInDays)
	sm := &model.SharedManifest{
		ID:        uuid.New().String(),
		OrgID:     orgID.String(),
		GroupID:   &gs,
		Token:     token,
		ExpiresAt: &expiresAt,
		IsActive:  true,
	}
	if pin != "" {
		sm.PinHash = &pin
	}
	if err := s.repo.CreateSharedManifest(ctx, sm); err != nil {
		return nil, err
	}
	return sm, nil
}

func (s *JamaahService) RevokeShare(ctx context.Context, orgID, groupID uuid.UUID) error {
	return s.repo.RevokeSharedManifest(ctx, groupID, orgID)
}

func (s *JamaahService) GetSharedManifest(ctx context.Context, token string) (*model.SharedManifest, error) {
	return s.repo.GetSharedManifestByToken(ctx, token)
}

func (s *JamaahService) GetRoomingSummary(ctx context.Context, orgID uuid.UUID, groupID *uuid.UUID) (*model.RoomingSummary, error) {
	return s.repo.GetRoomingSummary(ctx, orgID, groupID)
}
