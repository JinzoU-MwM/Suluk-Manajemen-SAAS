package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jamaah-in/v2/internal/auth/model"
)

func (r *AuthRepo) CreateTicket(ctx context.Context, t *model.Ticket) error {
	query := `
		INSERT INTO tickets (id, org_id, user_id, subject, message, priority, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		t.ID, t.OrgID, t.UserID, t.Subject, t.Message, t.Priority, t.Status,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *AuthRepo) ListTickets(ctx context.Context, orgID uuid.UUID, status string, limit, offset int) ([]model.Ticket, int, error) {
	countQuery := "SELECT COUNT(*) FROM tickets WHERE org_id = $1"
	args := []any{orgID}
	if status != "" {
		countQuery += " AND status = $2"
		args = append(args, status)
	}
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count tickets: %w", err)
	}

	query := `SELECT id, org_id, user_id, subject, message, priority, status, created_at, updated_at
		FROM tickets WHERE org_id = $1`
	queryArgs := []any{orgID}
	argIdx := 2
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		queryArgs = append(queryArgs, status)
		argIdx++
	}
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.OrgID, &t.UserID, &t.Subject, &t.Message,
			&t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

func (r *AuthRepo) GetTicketByID(ctx context.Context, ticketID, orgID uuid.UUID) (*model.Ticket, error) {
	query := `SELECT id, org_id, user_id, subject, message, priority, status, created_at, updated_at
		FROM tickets WHERE id = $1 AND org_id = $2`
	var t model.Ticket
	err := r.pool.QueryRow(ctx, query, ticketID, orgID).Scan(
		&t.ID, &t.OrgID, &t.UserID, &t.Subject, &t.Message,
		&t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return &t, nil
}

func (r *AuthRepo) UpdateTicketStatus(ctx context.Context, ticketID, orgID uuid.UUID, status string) error {
	query := `UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3`
	_, err := r.pool.Exec(ctx, query, status, ticketID, orgID)
	return err
}

func (r *AuthRepo) AddTicketMessage(ctx context.Context, msg *model.TicketMessage) error {
	query := `
		INSERT INTO ticket_messages (id, ticket_id, user_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	return r.pool.QueryRow(ctx, query, msg.ID, msg.TicketID, msg.UserID, msg.Content).Scan(&msg.CreatedAt)
}

func (r *AuthRepo) ListTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]model.TicketMessage, error) {
	query := `SELECT id, ticket_id, user_id, content, created_at
		FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []model.TicketMessage
	for rows.Next() {
		var m model.TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *AuthRepo) GetUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Ticket, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user tickets: %w", err)
	}

	query := `SELECT id, org_id, user_id, subject, message, priority, status, created_at, updated_at
		FROM tickets WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user tickets: %w", err)
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.OrgID, &t.UserID, &t.Subject, &t.Message,
			&t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

func (r *AuthRepo) GetAllTickets(ctx context.Context, status string, limit, offset int) ([]model.Ticket, int, error) {
	countQuery := "SELECT COUNT(*) FROM tickets"
	args := []any{}
	if status != "" {
		countQuery += " WHERE status = $1"
		args = append(args, status)
	}
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count all tickets: %w", err)
	}

	query := `SELECT id, org_id, user_id, subject, message, priority, status, created_at, updated_at FROM tickets`
	queryArgs := []any{}
	argIdx := 1
	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIdx)
		queryArgs = append(queryArgs, status)
		argIdx++
	}
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list all tickets: %w", err)
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.OrgID, &t.UserID, &t.Subject, &t.Message,
			&t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

func (r *AuthRepo) GetUnreadTicketCount(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE status NOT IN ('closed', 'resolved')`).Scan(&count)
	return count, err
}

func (r *AuthRepo) DeleteTicket(ctx context.Context, ticketID, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tickets WHERE id = $1 AND org_id = $2`, ticketID, orgID)
	return err
}

func (r *AuthRepo) TicketHasAccess(ctx context.Context, ticketID, orgID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND org_id = $2)`, ticketID, orgID).Scan(&exists)
	return exists, err
}
