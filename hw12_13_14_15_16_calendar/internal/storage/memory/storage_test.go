package memorystorage

import (
	"errors"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateEvent проверяет создание события
func TestCreateEvent(t *testing.T) {
	mem := New()

	event := storagecontracts.Event{
		Title:       "Тестовая встреча",
		From:        time.Now().Add(1 * time.Hour),
		To:          time.Now().Add(2 * time.Hour),
		Description: "Описание встречи",
		OwnerID:     "user1",
	}

	createdEvent, err := mem.Create(event)
	require.NoError(t, err)
	assert.NotEmpty(t, createdEvent.ID)
	assert.Equal(t, event.Title, createdEvent.Title)
	assert.Equal(t, event.OwnerID, createdEvent.OwnerID)
	assert.WithinDuration(t, event.From, createdEvent.From, time.Second)
}

// TestCreateEventValidation проверяет валидацию при создании
func TestCreateEventValidation(t *testing.T) {
	mem := New()
	now := time.Now()

	testCases := []struct {
		name      string
		event     storagecontracts.Event
		expectErr error
	}{
		{
			name: "Некорректная длительность - начало после окончания",
			event: storagecontracts.Event{
				Title:   "Некорректное событие",
				From:    now.Add(2 * time.Hour),
				To:      now.Add(1 * time.Hour),
				OwnerID: "user1",
			},
			expectErr: storagecontracts.ErrInvalidDuration,
		},
		{
			name: "Некорректная длительность - начало равно окончанию",
			event: storagecontracts.Event{
				Title:   "Некорректное событие",
				From:    now.Add(1 * time.Hour),
				To:      now.Add(1 * time.Hour),
				OwnerID: "user1",
			},
			expectErr: storagecontracts.ErrInvalidDuration,
		},
		{
			name: "Пустой заголовок",
			event: storagecontracts.Event{
				Title:   "",
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "user1",
			},
			expectErr: storagecontracts.ErrMissingRequiredFields,
		},
		{
			name: "Пустой OwnerID",
			event: storagecontracts.Event{
				Title:   "Встреча",
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "",
			},
			expectErr: storagecontracts.ErrMissingRequiredFields,
		},
		{
			name: "Событие в прошлом",
			event: storagecontracts.Event{
				Title:   "Прошлое событие",
				From:    now.Add(-2 * time.Hour),
				To:      now.Add(-1 * time.Hour),
				OwnerID: "user1",
			},
			expectErr: storagecontracts.ErrPastEvent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mem.Create(tc.event)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectErr.Error())
		})
	}
}

// TestCreateEventTimeOverlap проверяет пересечение событий по времени
func TestCreateEventTimeOverlap(t *testing.T) {
	mem := New()
	now := time.Now()

	// Создаем первое событие
	event1 := storagecontracts.Event{
		Title:   "Первая встреча",
		From:    now.Add(1 * time.Hour),
		To:      now.Add(2 * time.Hour),
		OwnerID: "user1",
	}

	_, err := mem.Create(event1)
	require.NoError(t, err)

	// Пытаемся создать пересекающееся событие
	overlappingEvent := storagecontracts.Event{
		Title:   "Пересекающаяся встреча",
		From:    now.Add(90 * time.Minute),
		To:      now.Add(150 * time.Minute),
		OwnerID: "user1",
	}

	_, err = mem.Create(overlappingEvent)
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrDateBusy, err)

	// Не пересекающееся событие у того же пользователя должно создаться
	nonOverlappingEvent := storagecontracts.Event{
		Title:   "Не пересекающаяся встреча",
		From:    now.Add(3 * time.Hour),
		To:      now.Add(4 * time.Hour),
		OwnerID: "user1",
	}

	_, err = mem.Create(nonOverlappingEvent)
	assert.NoError(t, err)

	// Пересекающееся событие у другого пользователя должно создаться
	otherUserEvent := storagecontracts.Event{
		Title:   "Встреча другого пользователя",
		From:    now.Add(90 * time.Minute),
		To:      now.Add(150 * time.Minute),
		OwnerID: "user2",
	}

	_, err = mem.Create(otherUserEvent)
	assert.NoError(t, err)
}

// TestUpdateEvent проверяет обновление события
func TestUpdateEvent(t *testing.T) {
	mem := New()
	now := time.Now()

	// Создаем событие
	event := storagecontracts.Event{
		Title:   "Исходная встреча",
		From:    now.Add(1 * time.Hour),
		To:      now.Add(2 * time.Hour),
		OwnerID: "user1",
	}

	createdEvent, err := mem.Create(event)
	require.NoError(t, err)

	// Обновляем событие
	updatedEvent := storagecontracts.Event{
		Title:   "Обновленная встреча",
		From:    now.Add(3 * time.Hour),
		To:      now.Add(4 * time.Hour),
		OwnerID: "user1", // Должен остаться прежним
	}

	result, err := mem.Update(createdEvent.ID, updatedEvent)
	require.NoError(t, err)

	assert.Equal(t, createdEvent.ID, result.ID)
	assert.Equal(t, "Обновленная встреча", result.Title)
	assert.WithinDuration(t, updatedEvent.From, result.From, time.Second)
	assert.Equal(t, "user1", result.OwnerID) // OwnerID не должен меняться
}

// TestUpdateNonExistentEvent проверяет обновление несуществующего события
func TestUpdateNonExistentEvent(t *testing.T) {
	mem := New()

	event := storagecontracts.Event{
		Title:   "Встреча",
		From:    time.Now().Add(1 * time.Hour),
		To:      time.Now().Add(2 * time.Hour),
		OwnerID: "user1",
	}

	_, err := mem.Update("non-existent-id", event)
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrEventNotFound, err)
}

// TestUpdateEventOverlap проверяет пересечение при обновлении
func TestUpdateEventOverlap(t *testing.T) {
	mem := New()
	now := time.Now()

	// Создаем два события
	event1 := storagecontracts.Event{
		Title:   "Первая встреча",
		From:    now.Add(1 * time.Hour),
		To:      now.Add(2 * time.Hour),
		OwnerID: "user1",
	}

	event2 := storagecontracts.Event{
		Title:   "Вторая встреча",
		From:    now.Add(3 * time.Hour),
		To:      now.Add(4 * time.Hour),
		OwnerID: "user1",
	}

	createdEvent1, err := mem.Create(event1)
	require.NoError(t, err)

	createdEvent2, err := mem.Create(event2)
	require.NoError(t, err)

	// Пытаемся обновить второе событие так, чтобы оно пересеклось с первым
	updatedEvent2 := storagecontracts.Event{
		Title:   "Обновленная вторая встреча",
		From:    now.Add(90 * time.Minute), // Пересекается с первым событием
		To:      now.Add(150 * time.Minute),
		OwnerID: "user1",
	}

	_, err = mem.Update(createdEvent2.ID, updatedEvent2)
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrDateBusy, err)

	// Проверяем, что исходные события не изменились
	retrievedEvent1, err := mem.GetEventByID(createdEvent1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Первая встреча", retrievedEvent1.Title)

	retrievedEvent2, err := mem.GetEventByID(createdEvent2.ID)
	require.NoError(t, err)
	assert.Equal(t, "Вторая встреча", retrievedEvent2.Title)
}

// TestDeleteEvent проверяет удаление события
func TestDeleteEvent(t *testing.T) {
	mem := New()

	event := storagecontracts.Event{
		Title:   "Встреча для удаления",
		From:    time.Now().Add(1 * time.Hour),
		To:      time.Now().Add(2 * time.Hour),
		OwnerID: "user1",
	}

	createdEvent, err := mem.Create(event)
	require.NoError(t, err)

	// Удаляем событие
	err = mem.Delete(createdEvent.ID)
	assert.NoError(t, err)

	// Проверяем, что событие удалено
	_, err = mem.GetEventByID(createdEvent.ID)
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrEventNotFound, err)
}

// TestDeleteNonExistentEvent проверяет удаление несуществующего события
func TestDeleteNonExistentEvent(t *testing.T) {
	mem := New()

	err := mem.Delete("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrEventNotFound, err)
}

// TestGetEventsForDay проверяет получение событий на день
func TestGetEventsForDay(t *testing.T) {
	mem := New()
	now := time.Now()
	today := storageutils.NormalizeDate(now)
	today = today.AddDate(0, 0, 1)
	tomorrow := today.AddDate(0, 0, 1)

	// Создаем события на сегодня
	event1 := storagecontracts.Event{
		Title:   "Утренняя встреча",
		From:    today.Add(9 * time.Hour),
		To:      today.Add(10 * time.Hour),
		OwnerID: "user1",
	}

	event2 := storagecontracts.Event{
		Title:   "Вечерняя встреча",
		From:    today.Add(18 * time.Hour),
		To:      today.Add(19 * time.Hour),
		OwnerID: "user1",
	}

	// Создаем событие на завтра
	event3 := storagecontracts.Event{
		Title:   "Завтрашняя встреча",
		From:    tomorrow.Add(14 * time.Hour),
		To:      tomorrow.Add(15 * time.Hour),
		OwnerID: "user1",
	}

	_, err := mem.Create(event1)
	require.NoError(t, err)

	_, err = mem.Create(event2)
	require.NoError(t, err)

	_, err = mem.Create(event3)
	require.NoError(t, err)

	// Получаем события на сегодня
	events := mem.GetEventsForDay(today)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	// Проверяем, что получены правильные события
	titles := make([]string, len(events))
	for i, e := range events {
		titles[i] = e.Title
	}
	assert.Contains(t, titles, "Утренняя встреча")
	assert.Contains(t, titles, "Вечерняя встреча")
	assert.NotContains(t, titles, "Завтрашняя встреча")

	// Получаем события на завтра
	events = mem.GetEventsForDay(tomorrow)
	assert.Len(t, events, 1)
	assert.Equal(t, "Завтрашняя встреча", events[0].Title)
}

// TestGetEventsForWeek проверяет получение событий на неделю
func TestGetEventsForWeek(t *testing.T) {
	mem := New()
	now := time.Now()
	startOfWeek := storageutils.NormalizeDate(now.AddDate(0, 0, 1))

	// Создаем события на разные дни недели
	events := []storagecontracts.Event{
		{
			Title:   "Понедельник",
			From:    startOfWeek.Add(10 * time.Hour),
			To:      startOfWeek.Add(11 * time.Hour),
			OwnerID: "user1",
		},
		{
			Title:   "Среда",
			From:    startOfWeek.AddDate(0, 0, 2).Add(14 * time.Hour),
			To:      startOfWeek.AddDate(0, 0, 2).Add(15 * time.Hour),
			OwnerID: "user1",
		},
		{
			Title:   "Пятница",
			From:    startOfWeek.AddDate(0, 0, 4).Add(16 * time.Hour),
			To:      startOfWeek.AddDate(0, 0, 4).Add(17 * time.Hour),
			OwnerID: "user1",
		},
	}

	for _, event := range events {
		_, err := mem.Create(event)
		require.NoError(t, err)
	}

	// Получаем события на неделю
	weekEvents := mem.GetEventsForWeek(startOfWeek)
	assert.Len(t, weekEvents, 3)

	// Проверяем, что события следующей недели не попали в результат
	nextWeekEvent := storagecontracts.Event{
		Title:   "Следующая неделя",
		From:    startOfWeek.AddDate(0, 0, 7).Add(10 * time.Hour),
		To:      startOfWeek.AddDate(0, 0, 7).Add(11 * time.Hour),
		OwnerID: "user1",
	}

	_, err := mem.Create(nextWeekEvent)
	require.NoError(t, err)

	weekEvents = mem.GetEventsForWeek(startOfWeek)
	assert.Len(t, weekEvents, 3) // Все еще 3 события
}

// TestGetEventsForMonth проверяет получение событий на месяц
func TestGetEventsForMonth(t *testing.T) {
	mem := New()
	now := time.Now()
	startOfMonth := storageutils.NormalizeDate(now.AddDate(0, 1, 0))

	// Создаем события на разные дни месяца
	events := []storagecontracts.Event{
		{
			Title:   "Первое число",
			From:    startOfMonth.Add(10 * time.Hour),
			To:      startOfMonth.Add(11 * time.Hour),
			OwnerID: "user1",
		},
		{
			Title:   "15 число",
			From:    startOfMonth.AddDate(0, 0, 14).Add(14 * time.Hour),
			To:      startOfMonth.AddDate(0, 0, 14).Add(15 * time.Hour),
			OwnerID: "user1",
		},
		{
			Title:   "Последний день месяца",
			From:    startOfMonth.AddDate(0, 1, -1).Add(16 * time.Hour),
			To:      startOfMonth.AddDate(0, 1, -1).Add(17 * time.Hour),
			OwnerID: "user1",
		},
	}

	for _, event := range events {
		_, err := mem.Create(event)
		require.NoError(t, err)
	}

	// Получаем события на месяц
	monthEvents := mem.GetEventsForMonth(startOfMonth)
	assert.Len(t, monthEvents, 3)

	// Создаем событие на следующий месяц
	nextMonthEvent := storagecontracts.Event{
		Title:   "Следующий месяц",
		From:    startOfMonth.AddDate(0, 1, 0).Add(10 * time.Hour),
		To:      startOfMonth.AddDate(0, 1, 0).Add(11 * time.Hour),
		OwnerID: "user1",
	}

	_, err := mem.Create(nextMonthEvent)
	require.NoError(t, err)

	// Проверяем, что событие следующего месяца не попало в результат
	monthEvents = mem.GetEventsForMonth(startOfMonth)
	require.NoError(t, err)
	assert.Len(t, monthEvents, 3)
}

// TestGetEventByID проверяет получение события по ID
func TestGetEventByID(t *testing.T) {
	mem := New()

	event := storagecontracts.Event{
		Title:   "Тестовая встреча",
		From:    time.Now().Add(1 * time.Hour),
		To:      time.Now().Add(2 * time.Hour),
		OwnerID: "user1",
	}

	createdEvent, err := mem.Create(event)
	require.NoError(t, err)

	// Получаем событие по ID
	retrievedEvent, err := mem.GetEventByID(createdEvent.ID)
	require.NoError(t, err)
	assert.Equal(t, createdEvent.ID, retrievedEvent.ID)
	assert.Equal(t, event.Title, retrievedEvent.Title)
	assert.Equal(t, event.OwnerID, retrievedEvent.OwnerID)
}

// TestGetNonExistentEventByID проверяет получение несуществующего события
func TestGetNonExistentEventByID(t *testing.T) {
	mem := New()

	_, err := mem.GetEventByID("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, storagecontracts.ErrEventNotFound, err)
}

// TestConcurrentAccess проверяет конкурентный доступ
func TestConcurrentAccess(t *testing.T) {
	mem := New()
	now := time.Now().AddDate(0, 0, 1)

	// Количество горутин
	numGoroutines := 10
	numEventsPerGoroutine := 100

	var wg sync.WaitGroup
	errs := make(chan error, numGoroutines*numEventsPerGoroutine)

	// Запускаем горутины
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(OwnerID string) {
			defer wg.Done()

			for j := 0; j < numEventsPerGoroutine; j++ {
				event := storagecontracts.Event{
					Title:   "Встреча",
					From:    now.Add(time.Duration(j) * time.Hour),
					To:      now.Add(time.Duration(j+1) * time.Hour),
					OwnerID: OwnerID,
				}

				_, err := mem.Create(event)
				if err != nil && !errors.Is(err, storagecontracts.ErrDateBusy) {
					// Игнорируем ошибки пересечения (ожидаемы при конкурентном создании)
					errs <- err
				}
			}
		}(string(rune('A' + i))) // Создаем разные OwnerID
	}

	wg.Wait()
	close(errs)

	// Проверяем, что не было ошибок кроме пересечений
	for err := range errs {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
}

// TestEventWithMultiDayDuration проверяет события с длительностью в несколько дней
func TestEventWithMultiDayDuration(t *testing.T) {
	mem := New()
	now := time.Now()

	// Создаем событие, которое длится 3 дня
	multiDayEvent := storagecontracts.Event{
		Title:   "Многодневное событие",
		From:    now.Add(24 * time.Hour), // Завтра
		To:      now.Add(72 * time.Hour), // Послезавтра + 1 день
		OwnerID: "user1",
	}

	createdEvent, err := mem.Create(multiDayEvent)
	require.NoError(t, err)

	// Проверяем, что событие есть во всех трех днях
	day1 := storageutils.NormalizeDate(createdEvent.From)
	day2 := day1.AddDate(0, 0, 1)
	day3 := day1.AddDate(0, 0, 2)

	eventsDay1 := mem.GetEventsForDay(day1)
	assert.Len(t, eventsDay1, 1)

	eventsDay2 := mem.GetEventsForDay(day2)
	assert.Len(t, eventsDay2, 1)

	eventsDay3 := mem.GetEventsForDay(day3)
	assert.Len(t, eventsDay3, 1)
}

// TestNormalizeDate проверяет нормализацию даты
func TestNormalizeDate(t *testing.T) {
	now := time.Now()
	normalized := storageutils.NormalizeDate(now)

	assert.Equal(t, now.Year(), normalized.Year())
	assert.Equal(t, now.Month(), normalized.Month())
	assert.Equal(t, now.Day(), normalized.Day())
	assert.Equal(t, 0, normalized.Hour())
	assert.Equal(t, 0, normalized.Minute())
	assert.Equal(t, 0, normalized.Second())
	assert.Equal(t, 0, normalized.Nanosecond())
}

// TestEventsOverlap проверяет функцию проверки пересечения событий
func TestEventsOverlap(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name    string
		event1  storagecontracts.Event
		event2  storagecontracts.Event
		overlap bool
	}{
		{
			name: "События не пересекаются",
			event1: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "user1",
			},
			event2: storagecontracts.Event{
				From:    now.Add(3 * time.Hour),
				To:      now.Add(4 * time.Hour),
				OwnerID: "user1",
			},
			overlap: false,
		},
		{
			name: "События пересекаются частично",
			event1: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(3 * time.Hour),
				OwnerID: "user1",
			},
			event2: storagecontracts.Event{
				From:    now.Add(2 * time.Hour),
				To:      now.Add(4 * time.Hour),
				OwnerID: "user1",
			},
			overlap: true,
		},
		{
			name: "Событие полностью внутри другого",
			event1: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(4 * time.Hour),
				OwnerID: "user1",
			},
			event2: storagecontracts.Event{
				From:    now.Add(2 * time.Hour),
				To:      now.Add(3 * time.Hour),
				OwnerID: "user1",
			},
			overlap: true,
		},
		{
			name: "События касаются концами",
			event1: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "user1",
			},
			event2: storagecontracts.Event{
				From:    now.Add(2 * time.Hour),
				To:      now.Add(3 * time.Hour),
				OwnerID: "user1",
			},
			overlap: false, // Касание не считается пересечением
		},
		{
			name: "Разные пользователи - нет пересечения",
			event1: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "user1",
			},
			event2: storagecontracts.Event{
				From:    now.Add(1 * time.Hour),
				To:      now.Add(2 * time.Hour),
				OwnerID: "user2", // Разный пользователь
			},
			overlap: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := eventsOverlap(tc.event1, tc.event2)
			assert.Equal(t, tc.overlap, result)
		})
	}
}
