package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
	storageutils "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/utils"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib" // justifying comment
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db      *sqlx.DB
	configs configuration.DbConf
}

func New(configs configuration.DbConf) *Storage {
	return &Storage{
		configs: configs,
		db:      nil,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	var err error
	if s.db, err = sqlx.Open("pgx", makeDsnFromConfig(s.configs)); err != nil {
		return fmt.Errorf("enable to connect to the database: %w", err)
	}

	// Проверяем соединение
	if err = s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("enable to ping database: %w", err)
	}

	// Устанавливаем настройки пула соединений
	s.db.SetMaxOpenConns(s.configs.MaxOpenConn)
	s.db.SetMaxIdleConns(s.configs.MaxIdleConn)
	s.db.SetConnMaxLifetime(s.configs.MaxLifetimeConn * time.Minute) //nolint:durationcheck

	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) Create(event storagecontracts.Event) (storagecontracts.Event, error) {
	// Генерация ID
	event.ID = uuid.New().String()

	// Начинаем транзакцию
	tx, err := s.db.Beginx()
	if err != nil {
		return storagecontracts.Event{}, fmt.Errorf("transaction begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Проверяем, что время начала не в прошлом
	if event.From.Before(time.Now()) {
		return storagecontracts.Event{}, storagecontracts.ErrPastEvent
	}

	// Проверяем пересечение с существующими событиями
	if err := s.checkTimeOverlap(tx, event); err != nil {
		return storagecontracts.Event{}, err
	}

	// Вставляем событие
	query := `
		INSERT INTO events (id, title, start_time, end_time, description, owner_id, notify_time)
		VALUES (:id, :title, :start_time, :end_time, :description, :owner_id, :notify_time)
		RETURNING id, title, start_time, end_time, description, owner_id, notify_time
	`

	var result storagecontracts.Event
	rows, err := tx.NamedQuery(
		query,
		map[string]interface{}{
			"id":          event.ID,
			"title":       event.Title,
			"start_time":  event.From,
			"end_time":    event.To,
			"description": event.Description,
			"owner_id":    event.OwnerID,
			"notify_time": event.Notify,
		},
	)
	if err != nil {
		return storagecontracts.Event{}, fmt.Errorf("creating event: %w", err)
	}

	if err := rows.StructScan(&result); err != nil {
		return storagecontracts.Event{}, err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return storagecontracts.Event{}, fmt.Errorf("transaction commit: %w", err)
	}

	return result, nil
}

func (s *Storage) Update(id string, event storagecontracts.Event) (storagecontracts.Event, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return storagecontracts.Event{}, fmt.Errorf("transaction begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Получаем существующее событие для проверки
	_, err = s.getEventByID(tx, id)
	if err != nil {
		return storagecontracts.Event{}, err
	}

	// Проверяем пересечение, исключая текущее событие
	if err := s.checkTimeOverlapExcluding(tx, event); err != nil {
		return storagecontracts.Event{}, err
	}

	// Обновляем событие
	query := `
		UPDATE events 
		SET title = :title, 
		    start_time = :start_time, 
		    end_time = :end_time, 
		    description = :description, 
		    notify_time = :notify_time,
		    updated_at = :updated_at
		WHERE id = :id
		RETURNING id, title, start_time, end_time, description, owner_id, notify_time
	`

	var result storagecontracts.Event
	rows, err := tx.NamedQuery(
		query,
		map[string]interface{}{
			"title":       event.Title,
			"start_time":  event.From,
			"end_time":    event.To,
			"description": event.Description,
			"notify_time": event.Notify,
			"updated_at":  time.Now(),
		},
	)
	if err != nil {
		return storagecontracts.Event{}, fmt.Errorf("event update: %w", err)
	}

	if err := rows.StructScan(&result); err != nil {
		return storagecontracts.Event{}, err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return storagecontracts.Event{}, fmt.Errorf("transaction commit: %w", err)
	}

	return result, nil
}

func (s *Storage) Delete(id string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("transaction begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := "DELETE FROM events WHERE id = :id"
	result, err := s.db.NamedExec(
		query,
		map[string]interface{}{
			"id": id,
		},
	)
	if err != nil {
		return fmt.Errorf("event delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("obtain deleted rows count: %w", err)
	}

	if rowsAffected == 0 {
		return storagecontracts.ErrEventNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit: %w", err)
	}

	return nil
}

// GetEventsForDay возвращает события на конкретный день.
func (s *Storage) GetEventsForDay(day time.Time) ([]storagecontracts.Event, error) {
	start := storageutils.NormalizeDate(day)
	end := start.AddDate(0, 0, 1)

	return s.getEventsForPeriod(start, end)
}

// GetEventsForWeek возвращает события на неделю.
func (s *Storage) GetEventsForWeek(startOfWeek time.Time) ([]storagecontracts.Event, error) {
	start := storageutils.NormalizeDate(startOfWeek)
	end := start.AddDate(0, 0, 7)

	return s.getEventsForPeriod(start, end)
}

// GetEventsForMonth возвращает события на месяц.
func (s *Storage) GetEventsForMonth(startOfMonth time.Time) ([]storagecontracts.Event, error) {
	// Нормализуем начало месяца
	start := storageutils.NormalizeDate(startOfMonth)
	end := start.AddDate(0, 1, 0)

	return s.getEventsForPeriod(start, end)
}

// checkTimeOverlap проверяет пересечение по времени.
func (s *Storage) checkTimeOverlap(tx *sqlx.Tx, event storagecontracts.Event) error {
	query := `
		SELECT COUNT(*) 
		FROM events 
		WHERE owner_id = :owner_id 
		  AND (
			(start_time < :to AND end_time > :from) OR  -- Существующее событие охватывает новое
			(start_time >= :to AND start_time < :from) OR  -- Начало нового внутри существующего
			(end_time > :to AND end_time <= :from) OR  -- Конец нового внутри существующего
			(start_time >= :to AND end_time <= :from)  -- Новое событие охватывает существующее
		  )
	`

	var count int
	rows, err := tx.NamedQuery(
		query,
		map[string]interface{}{
			"owner_id": event.OwnerID,
			"from":     event.From,
			"to":       event.To,
		},
	)
	if err != nil {
		return fmt.Errorf("time overlap check: %w", err)
	}

	if err := rows.Scan(count); err != nil {
		return err
	}

	if count > 0 {
		return storagecontracts.ErrDateBusy
	}

	return nil
}

// checkTimeOverlapExcluding проверяет пересечение, исключая указанное событие.
func (s *Storage) checkTimeOverlapExcluding(tx *sqlx.Tx, event storagecontracts.Event) error {
	query := `
		SELECT COUNT(*) 
		FROM events 
		WHERE owner_id = :owner_id 
		  AND id != :id
		  AND (
			(start_time < :to AND end_time > :from) OR  -- Существующее событие охватывает новое
			(start_time >= :to AND start_time < :from) OR  -- Начало нового внутри существующего
			(end_time > :to AND end_time <= :from) OR  -- Конец нового внутри существующего
			(start_time >= :to AND end_time <= :from)  -- Новое событие охватывает существующее
		  )
	`

	var count int
	rows, err := tx.NamedQuery(
		query,
		map[string]interface{}{
			"owner_id": event.OwnerID,
			"from":     event.From,
			"to":       event.To,
			"id":       event.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("time overlap check: %w", err)
	}

	if err := rows.Scan(count); err != nil {
		return err
	}

	if count > 0 {
		return storagecontracts.ErrDateBusy
	}

	return nil
}

// getEventByID возвращает событие по ID.
func (s *Storage) getEventByID(tx *sqlx.Tx, id string) (storagecontracts.Event, error) {
	query := `
		SELECT id, title, start_time, end_time, description, owner_id, notify_time
		FROM events
		WHERE id = :id
	`

	var event storagecontracts.Event
	rows, err := tx.NamedQuery(
		query,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storagecontracts.Event{}, storagecontracts.ErrEventNotFound
		}
		return storagecontracts.Event{}, fmt.Errorf("obtain event by id: %w", err)
	}

	if err := rows.StructScan(&event); err != nil {
		return storagecontracts.Event{}, err
	}

	return event, nil
}

// getEventsForPeriod возвращает список событий за переданный период.
func (s *Storage) getEventsForPeriod(start, end time.Time) ([]storagecontracts.Event, error) {
	query := `
		SELECT id, title, start_time, end_time, description, owner_id, notify_time
		FROM events
		WHERE start_time >= :start_time AND start_time < :end_time
		ORDER BY start_time
	`

	rows, err := s.db.NamedQuery(
		query,
		map[string]interface{}{
			"start_time": start,
			"end_time":   end,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("obtain events for period: %w", err)
	}

	events := make([]storagecontracts.Event, 0)
	for rows.Next() {
		var event storagecontracts.Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, fmt.Errorf("scan event row: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
