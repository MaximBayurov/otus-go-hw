package httphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
)

const contentType = "application/json"

type Handler func(http.ResponseWriter, *http.Request)

func CreateEvent(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		if contentType != r.Header.Get("Content-Type") {
			http.Error(w, fmt.Sprintf("Content-Type must be %s", contentType), http.StatusUnsupportedMediaType)
			return
		}

		var event storagecontracts.Event

		// Декодируем JSON тело запроса в структуру
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() // запрещаем неизвестные поля

		err := decoder.Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Закрываем тело запроса
		defer func() {
			if err := r.Body.Close(); err != nil {
				logger.Error(err.Error())
			}
		}()

		if event, err = app.CreateEvent(
			r.Context(),
			event.Title,
			event.From,
			event.To,
			event.Description,
			event.OwnerID,
			event.Notify,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(map[string]interface{}{
			"event": event,
		}); err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func UpdateEvent(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPatch {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		if contentType != r.Header.Get("Content-Type") {
			http.Error(w, fmt.Sprintf("Content-Type must be %s", contentType), http.StatusUnsupportedMediaType)
			return
		}

		var event storagecontracts.Event

		// Декодируем JSON тело запроса в структуру
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() // запрещаем неизвестные поля

		err := decoder.Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Закрываем тело запроса
		defer func() {
			if err := r.Body.Close(); err != nil {
				logger.Error(err.Error())
			}
		}()

		id := r.PathValue("id")
		if event, err = app.UpdateEvent(
			r.Context(),
			id,
			event.Title,
			event.From,
			event.To,
			event.Description,
			event.Notify,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(map[string]interface{}{
			"event": event,
		}); err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func DeleteEvent(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем Content-Type
		if contentType != r.Header.Get("Content-Type") {
			http.Error(w, fmt.Sprintf("Content-Type must be %s", contentType), http.StatusUnsupportedMediaType)
			return
		}

		id := r.PathValue("id")
		if err := app.DeleteEvent(
			r.Context(),
			id,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		}); err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type getEventFunc func(ctx context.Context, day time.Time) ([]storagecontracts.Event, error)

func makeGetEventsHandler(
	paramName string,
	getEvents getEventFunc,
	logger contracts.Logger,
) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Парсинг даты-времени (RFC3339/ISO 8601)
		var datetimeStr string
		if datetimeStr = r.URL.Query().Get(paramName); datetimeStr == "" {
			http.Error(w, fmt.Sprintf("missing required GET param \"%s\"", paramName), http.StatusBadRequest)
			return
		}

		var day time.Time
		var err error
		if day, err = time.Parse(time.RFC3339, datetimeStr); err != nil {
			http.Error(w, "Invalid datetime format. Use RFC3339 (2006-01-02T15:04:05Z07:00)", http.StatusBadRequest)
			return
		}

		var events []storagecontracts.Event
		if events, err = getEvents(
			r.Context(),
			day,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
		}); err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func GetEventsForDay(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return makeGetEventsHandler(
		"from",
		func(ctx context.Context, day time.Time) ([]storagecontracts.Event, error) {
			return app.GetEventsForDay(ctx, day)
		},
		logger,
	)
}

func GetEventsForWeek(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return makeGetEventsHandler(
		"from",
		func(ctx context.Context, day time.Time) ([]storagecontracts.Event, error) {
			return app.GetEventsForWeek(ctx, day)
		},
		logger,
	)
}

func GetEventsForMonth(
	app contracts.Application,
	logger contracts.Logger,
) Handler {
	return makeGetEventsHandler(
		"from",
		func(ctx context.Context, day time.Time) ([]storagecontracts.Event, error) {
			return app.GetEventsForMonth(ctx, day)
		},
		logger,
	)
}
