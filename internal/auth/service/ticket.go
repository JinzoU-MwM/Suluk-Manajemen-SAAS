package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/auth/model"
)

func (s *AuthService) CreateTicket(ctx context.Context, orgID, userID uuid.UUID, req model.CreateTicketRequest) (*model.Ticket, error) {
	if req.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	if priority != "low" && priority != "medium" && priority != "high" && priority != "urgent" {
		return nil, fmt.Errorf("invalid priority: must be low, medium, high, or urgent")
	}

	ticket := &model.Ticket{
		ID:       uuid.New(),
		OrgID:    orgID,
		UserID:   userID,
		Subject:  req.Subject,
		Message:  req.Message,
		Priority: priority,
		Status:   "open",
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	_ = s.repo.CreateNotification(ctx, &model.Notification{
		ID:       uuid.New(),
		OrgID:    orgID,
		Severity: "info",
		Title:    "Tiket Baru: " + req.Subject,
		Message:  req.Message,
	})

	return ticket, nil
}

func (s *AuthService) ListTickets(ctx context.Context, orgID uuid.UUID, status string, page, pageSize int) ([]model.Ticket, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListTickets(ctx, orgID, status, pageSize, offset)
}

func (s *AuthService) GetTicketWithMessages(ctx context.Context, ticketID, orgID uuid.UUID) (*model.TicketWithMessages, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, nil
	}

	messages, err := s.repo.ListTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []model.TicketMessage{}
	}

	return &model.TicketWithMessages{
		Ticket:   *ticket,
		Messages: messages,
	}, nil
}

func (s *AuthService) AddTicketMessage(ctx context.Context, ticketID, orgID, userID uuid.UUID, req model.AddTicketMessageRequest) (*model.TicketMessage, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	hasAccess, err := s.repo.TicketHasAccess(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, fmt.Errorf("ticket not found")
	}

	msg := &model.TicketMessage{
		ID:       uuid.New(),
		TicketID: ticketID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := s.repo.AddTicketMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	_ = s.repo.UpdateTicketStatus(ctx, ticketID, orgID, "open")

	return msg, nil
}

func (s *AuthService) GetUserTickets(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Ticket, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetUserTickets(ctx, userID, pageSize, offset)
}

func (s *AuthService) GetAllTickets(ctx context.Context, status string, page, pageSize int) ([]model.Ticket, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAllTickets(ctx, status, pageSize, offset)
}

func (s *AuthService) UpdateTicketStatus(ctx context.Context, ticketID, orgID uuid.UUID, status string) error {
	valid := map[string]bool{"open": true, "in_progress": true, "resolved": true, "closed": true}
	if !valid[status] {
		return fmt.Errorf("invalid status")
	}
	return s.repo.UpdateTicketStatus(ctx, ticketID, orgID, status)
}

func (s *AuthService) DeleteTicket(ctx context.Context, ticketID, orgID uuid.UUID) error {
	return s.repo.DeleteTicket(ctx, ticketID, orgID)
}
