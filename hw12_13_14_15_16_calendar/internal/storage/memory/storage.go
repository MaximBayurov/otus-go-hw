package memorystorage

import (
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
	storageutils "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storagecontracts.Event // Основной словарь событий по ID
	byDate map[time.Time]map[string]struct{} // Индексация по дате
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storagecontracts.Event),
		byDate: make(map[time.Time]map[string]struct{}),
	}
}

func (s *Storage) Create(event storagecontracts.Event) (storagecontracts.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Валидация
	if err := validateEvent(event); err != nil {
		return storagecontracts.Event{}, err
	}

	// Проверка пересечений по времени
	if err := s.checkTimeOverlap(event); err != nil {
		return storagecontracts.Event{}, err
	}

	// Генерация ID
	event.ID = uuid.New().String()

	// Проверяем, что время начала не в прошлом
	if event.From.Before(time.Now()) {
		return storagecontracts.Event{}, storagecontracts.ErrPastEvent
	}

	// Сохраняем событие
	s.events[event.ID] = event

	// Обновляем индексы
	s.updateIndexes(event)

	return event, nil
}

func (s *Storage) Update(ID string, event storagecontracts.Event) (storagecontracts.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем существование события
	existingEvent, exists := s.events[ID]
	if !exists {
		return storagecontracts.Event{}, storagecontracts.ErrEventNotFound
	}

	// Копируем ID пользователя из существующего события
	event.OwnerID = existingEvent.OwnerID
	event.ID = ID

	// Валидация
	if err := validateEvent(event); err != nil {
		return storagecontracts.Event{}, err
	}

	// Проверяем пересечения, исключая текущее событие
	if err := s.checkTimeOverlapExcluding(event, ID); err != nil {
		return storagecontracts.Event{}, err
	}

	// Удаляем из индексов старое событие
	s.removeFromIndexes(existingEvent)

	// Сохраняем обновленное событие
	s.events[ID] = event

	// Обновляем индексы
	s.updateIndexes(event)

	return event, nil
}

func (s *Storage) Delete(ID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, exists := s.events[ID]
	if !exists {
		return storagecontracts.ErrEventNotFound
	}

	// Удаляем из индексов
	s.removeFromIndexes(event)

	// Удаляем из основного словаря
	delete(s.events, ID)

	return nil
}

// GetEventsForDay возвращает события на конкретный день
func (s *Storage) GetEventsForDay(day time.Time) ([]storagecontracts.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Нормализуем дату (убираем время)
	date := storageutils.NormalizeDate(day)

	// Получаем события на эту дату
	eventIDs, exists := s.byDate[date]
	if !exists {
		return []storagecontracts.Event{}, nil
	}

	events := make([]storagecontracts.Event, 0, len(eventIDs))
	for id := range eventIDs {
		if event, ok := s.events[id]; ok {
			events = append(events, event)
		}
	}

	return events, nil
}

// GetEventsForWeek возвращает события на неделю
func (s *Storage) GetEventsForWeek(startOfWeek time.Time) ([]storagecontracts.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Нормализуем дату начала недели
	startDate := storageutils.NormalizeDate(startOfWeek)
	endDate := startDate.AddDate(0, 0, 7)

	return s.getEventsForPeriod(startOfWeek, endDate)
}

// GetEventsForMonth возвращает события на месяц
func (s *Storage) GetEventsForMonth(startOfMonth time.Time) ([]storagecontracts.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Нормализуем дату начала месяца
	startDate := storageutils.NormalizeDate(startOfMonth)
	// Получаем последний день месяца
	nextMonth := startDate.AddDate(0, 1, 0)

	return s.getEventsForPeriod(startDate, nextMonth)
}

// GetEventByID возвращает событие по ID
func (s *Storage) GetEventByID(ID string) (storagecontracts.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[ID]
	if !exists {
		return storagecontracts.Event{}, storagecontracts.ErrEventNotFound
	}

	return event, nil
}

// getEventsForPeriod возвращает события за переданный период
func (s *Storage) getEventsForPeriod(startDate, endDate time.Time) ([]storagecontracts.Event, error) {
	events := make([]storagecontracts.Event, 0)

	// Итерируем по дням недели
	for current := startDate; current.Before(endDate); current = current.AddDate(0, 0, 1) {
		var eventIDs map[string]struct{}
		var exists bool
		if eventIDs, exists = s.byDate[current]; !exists {
			continue
		}
		for id := range eventIDs {
			if event, ok := s.events[id]; ok {
				events = append(events, event)
			}
		}
	}

	return events, nil
}

// checkTimeOverlap проверяет пересечение по времени
func (s *Storage) checkTimeOverlap(event storagecontracts.Event) error {
	// Получаем все события на дату начала
	date := storageutils.NormalizeDate(event.From)
	eventIDs, exists := s.byDate[date]
	if !exists {
		return nil
	}

	// Проверяем пересечение с каждым событием
	for id := range eventIDs {
		existingEvent, ok := s.events[id]
		if !ok {
			continue
		}

		// Проверяем пересечение временных интервалов
		if eventsOverlap(event, existingEvent) {
			return storagecontracts.ErrDateBusy
		}
	}

	return nil
}

// checkTimeOverlapExcluding проверяет пересечение, исключая указанное событие
func (s *Storage) checkTimeOverlapExcluding(event storagecontracts.Event, excludeID string) error {
	date := storageutils.NormalizeDate(event.From)
	eventIDs, exists := s.byDate[date]
	if !exists {
		return nil
	}

	for id := range eventIDs {
		if id == excludeID {
			continue
		}

		existingEvent, ok := s.events[id]
		if !ok {
			continue
		}

		if eventsOverlap(event, existingEvent) {
			return storagecontracts.ErrDateBusy
		}
	}

	return nil
}

// updateIndexes обновляет индексы для события
func (s *Storage) updateIndexes(event storagecontracts.Event) {
	// Индекс по датам (для всех дней, которые охватывает событие)
	startDate := storageutils.NormalizeDate(event.From)
	endDate := storageutils.NormalizeDate(event.To)

	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		if _, exists := s.byDate[current]; !exists {
			s.byDate[current] = make(map[string]struct{})
		}
		s.byDate[current][event.ID] = struct{}{}
	}
}

// removeFromIndexes удаляет событие из индексов
func (s *Storage) removeFromIndexes(event storagecontracts.Event) {
	// Удаляем из индекса по датам
	startDate := storageutils.NormalizeDate(event.From)
	endDate := storageutils.NormalizeDate(event.To)

	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		if dateEvents, exists := s.byDate[current]; exists {
			delete(dateEvents, event.ID)
			if len(dateEvents) == 0 {
				delete(s.byDate, current)
			}
		}
	}
}

// validateEvent проверяет корректность события
func validateEvent(event storagecontracts.Event) error {
	// Проверка длительности
	if event.From.After(event.To) || event.From.Equal(event.To) {
		return storagecontracts.ErrInvalidDuration
	}

	// Проверка обязательных полей
	if event.Title == "" || event.OwnerID == "" {
		return storagecontracts.ErrMissingRequiredFields
	}

	return nil
}
