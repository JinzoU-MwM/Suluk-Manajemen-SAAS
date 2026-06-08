package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

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

func (s *JamaahService) AutoRooming(ctx context.Context, orgID, groupID uuid.UUID) ([]model.Room, error) {
	_ = s.repo.DeleteRoomsByGroup(ctx, groupID, orgID)

	members, err := s.repo.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var rooms []model.Room
	roomNum := 1
	for i := 0; i < len(members); i += 2 {
		gs := groupID.String()
		room := &model.Room{
			ID:         uuid.New().String(),
			OrgID:      orgID.String(),
			GroupID:    &gs,
			RoomNumber: fmt.Sprintf("%d", roomNum),
			GenderType: "mixed",
			RoomType:   "double",
			Capacity:   2,
			IsActive:   true,
		}
		if err := s.repo.CreateRoom(ctx, room); err == nil {
			rooms = append(rooms, *room)
			roomID, _ := uuid.Parse(room.ID)
			s.repo.AssignMemberToRoom(ctx, orgID, roomID, members[i].MemberID.String())
			if i+1 < len(members) {
				s.repo.AssignMemberToRoom(ctx, orgID, roomID, members[i+1].MemberID.String())
			}
		}
		roomNum++
	}
	if rooms == nil {
		rooms = []model.Room{}
	}
	return rooms, nil
}

func (s *JamaahService) ClearAutoRooming(ctx context.Context, orgID, groupID uuid.UUID) error {
	return s.repo.DeleteRoomsByGroup(ctx, groupID, orgID)
}

func (s *JamaahService) AssignMemberToRoom(ctx context.Context, orgID, roomID uuid.UUID, memberID string) error {
	return s.repo.AssignMemberToRoom(ctx, orgID, roomID, memberID)
}

func (s *JamaahService) UnassignMember(ctx context.Context, roomID uuid.UUID, memberID string) error {
	return s.repo.UnassignMember(ctx, roomID, memberID)
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
