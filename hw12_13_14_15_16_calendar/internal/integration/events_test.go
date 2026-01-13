package integration

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventsAPI тестирует API событий
type TestEventsAPI struct {
	TestSuite
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционные тесты в режиме short")
	}

	suite.Run(t, new(TestEventsAPI))
}

func (s *TestEventsAPI) TestCreateEvent() {
	s.T().Run("Успешное создание события", func(t *testing.T) {
		userID := "test-user-1"
		eventReq := Event{
			Title:       "Тестовая встреча",
			StartTime:   time.Now().Add(2 * time.Hour),
			EndTime:     time.Now().Add(3 * time.Hour),
			Description: "Описание тестовой встречи",
			OwnerId:     userID,
			Notify:      time.Now().Add(1 * time.Hour).Add(30 * time.Minute),
		}

		event, err := s.CreateEvent(eventReq)
		require.NoError(t, err)
		require.NotNil(t, event)

		assert.NotEmpty(t, event.ID)
		assert.Equal(t, eventReq.Title, event.Title)
		assert.Equal(t, eventReq.Description, event.Description)
		assert.Equal(t, userID, event.OwnerId)
		assert.WithinDuration(t, eventReq.Notify, event.Notify, time.Second)
		assert.WithinDuration(t, eventReq.StartTime, event.StartTime, time.Second)
		assert.WithinDuration(t, eventReq.EndTime, event.EndTime, time.Second)
	})

	s.T().Run("Ошибка при некорректной длительности", func(t *testing.T) {
		userID := "test-user-3"
		eventReq := Event{
			Title:       "Некорректное событие",
			StartTime:   time.Now().Add(2 * time.Hour),
			EndTime:     time.Now().Add(1 * time.Hour), // end_time раньше start_time
			Description: "Некорректное событие описание",
			OwnerId:     userID,
			Notify:      time.Now().Add(30 * time.Minute),
		}

		jsonData, err := json.Marshal(eventReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", s.baseURL+"/events/create", bytes.NewReader(jsonData))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.apiClient.httpClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	s.T().Run("Конфликт времени событий", func(t *testing.T) {
		userID := "test-user-4"

		// Первое событие
		event1Req := Event{
			Title:       "Первая встреча",
			StartTime:   time.Now().Add(6 * time.Hour),
			EndTime:     time.Now().Add(7 * time.Hour),
			Description: "Некорректное событие описание",
			OwnerId:     userID,
			Notify:      time.Now().Add(30 * time.Minute),
		}

		event1, err := s.CreateEvent(event1Req)
		require.NoError(t, err)
		require.NotNil(t, event1)

		// Второе событие, пересекающееся по времени
		event2Req := Event{
			Title:       "Пересекающаяся встреча",
			StartTime:   time.Now().Add(6*time.Hour + 30*time.Minute),
			EndTime:     time.Now().Add(7*time.Hour + 30*time.Minute),
			Description: "Некорректное событие описание",
			OwnerId:     userID,
			Notify:      time.Now().Add(30 * time.Minute),
		}

		jsonData, err := json.Marshal(event2Req)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", s.baseURL+"/events/create", bytes.NewReader(jsonData))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		resp, err := s.apiClient.httpClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func (s *TestEventsAPI) TestGetEventsForPeriod() {
	s.T().Run("Получение событий на день", func(t *testing.T) {
		userID := "test-user-10"
		tomorrow := time.Now().AddDate(0, 0, 1)
		tomorrow = time.Date(
			tomorrow.Year(),
			tomorrow.Month(),
			tomorrow.Day(),
			0, 0, 0, 0,
			tomorrow.Location(),
		)

		// Создаем события на сегодня
		eventsTomorrow := []Event{
			{
				Title:       "Утренняя встреча",
				StartTime:   tomorrow.Add(9 * time.Hour),
				EndTime:     tomorrow.Add(10 * time.Hour),
				Description: "Некорректное событие описание",
				OwnerId:     userID,
				Notify:      tomorrow.Add(30 * time.Minute),
			},
			{
				Title:       "Вечерняя встреча",
				StartTime:   tomorrow.Add(18 * time.Hour),
				EndTime:     tomorrow.Add(19 * time.Hour),
				Description: "Некорректное событие описание",
				OwnerId:     userID,
				Notify:      tomorrow.Add(30 * time.Minute),
			},
		}

		// Создаем события
		for _, eventReq := range eventsTomorrow {
			_, err := s.CreateEvent(eventReq)
			require.NoError(t, err)
		}

		// Получаем события на сегодня
		req, err := http.NewRequest("GET", s.baseURL+"/events/list/daily", nil)
		require.NoError(t, err)

		dateStr := tomorrow.Format(time.RFC3339)
		q := req.URL.Query()
		q.Add("from", dateStr)
		req.URL.RawQuery = q.Encode()

		resp, err := s.apiClient.httpClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsResponse struct {
			Events []Event `json:"events"`
		}
		err = json.NewDecoder(resp.Body).Decode(&eventsResponse)
		require.NoError(t, err)

		assert.Len(t, eventsResponse.Events, 2)

		// Проверяем, что получены правильные события
		titles := make([]string, len(eventsResponse.Events))
		for i, event := range eventsResponse.Events {
			titles[i] = event.Title
		}
		assert.Contains(t, titles, "Утренняя встреча")
		assert.Contains(t, titles, "Вечерняя встреча")
	})

	s.T().Run("Получение событий на неделю", func(t *testing.T) {
		userID := "test-user-10"
		startOfWeek := time.Now()

		// Создаем события на разные дни недели
		events := []struct {
			title string
			day   int
		}{
			{"Понедельник", 0},
			{"Среда", 2},
			{"Пятница", 4},
		}

		for _, event := range events {
			eventReq := Event{
				Title:     event.title,
				StartTime: startOfWeek.Add(time.Duration(event.day)*24*time.Hour + 10*time.Hour),
				EndTime:   startOfWeek.Add(time.Duration(event.day)*24*time.Hour + 11*time.Hour),
				OwnerId:   userID,
			}

			_, err := s.CreateEvent(eventReq)
			require.NoError(t, err)
		}

		// Получаем события на неделю
		dateStr := startOfWeek.Format(time.RFC3339)
		req, err := http.NewRequest("GET", s.baseURL+"/events/list/weekly?from="+dateStr, nil)
		require.NoError(t, err)

		resp, err := s.apiClient.httpClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsResponse struct {
			Events []Event `json:"events"`
		}
		err = json.NewDecoder(resp.Body).Decode(&eventsResponse)
		require.NoError(t, err)

		assert.Len(t, eventsResponse.Events, 5)
	})

	s.T().Run("Пустой список событий", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 0, 30)
		dateStr := futureDate.Format(time.RFC3339)

		req, err := http.NewRequest("GET", s.baseURL+"/events/list/daily?from="+dateStr, nil)
		require.NoError(t, err)

		resp, err := s.apiClient.httpClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsResponse struct {
			Events []Event `json:"events"`
		}
		err = json.NewDecoder(resp.Body).Decode(&eventsResponse)
		require.NoError(t, err)

		assert.Empty(t, eventsResponse.Events)
	})
}
