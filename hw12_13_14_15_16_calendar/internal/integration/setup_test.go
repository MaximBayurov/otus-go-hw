package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite - базовый класс для всех интеграционных тестов.
type TestSuite struct {
	suite.Suite
	apiClient   *APIClient
	db          *sql.DB
	baseURL     string
	databaseDSN string
}

// APIClient - клиент для работы с API.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient создает новый API клиент.
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetupSuite выполняется перед всеми тестами.
func (s *TestSuite) SetupSuite() {
	// Читаем конфигурацию из переменных окружения
	s.baseURL = os.Getenv("API_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	s.databaseDSN = os.Getenv("DATABASE_DSN")
	if s.databaseDSN == "" {
		s.databaseDSN = "postgres://postgres:postgres@localhost:5432/calendar_test?sslmode=disable"
	}

	// Инициализируем клиент
	s.apiClient = NewAPIClient(s.baseURL)

	// Подключаемся к БД
	var err error
	s.db, err = sql.Open("postgres", s.databaseDSN)
	require.NoError(s.T(), err)

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = s.db.PingContext(ctx)
	require.NoError(s.T(), err)

	// Очищаем базу данных перед запуском тестов
	s.cleanDatabase()
}

// TearDownSuite выполняется после всех тестов.
func (s *TestSuite) TearDownSuite() {
	if s.db != nil {
		_ = s.db.Close()
	}
}

// SetupTest выполняется перед каждым тестом.
func (s *TestSuite) SetupTest() {
	// Очищаем базу данных перед каждым тестом
	s.cleanDatabase()
}

// cleanDatabase очищает базу данных.
func (s *TestSuite) cleanDatabase() {
	tables := []string{
		"events",
	}

	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table)) //nolint:noctx
		if err != nil {
			// Таблица может не существовать, игнорируем ошибку
			log.Printf("Warning: не удалось очистить таблицу %s: %v", table, err)
		}
	}
}

// CreateEvent создает событие через API.
func (s *TestSuite) CreateEvent(event Event) (*Event, error) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"PUT",
		s.baseURL+"/events/create",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.apiClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errorResp["error"])
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var eventResp EventResp
	if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
		return nil, err
	}

	return &eventResp.Event, nil
}

// Event - запрос на создание события.
type Event struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"from"`
	EndTime     time.Time `json:"to"`
	Description string    `json:"description"`
	OwnerID     string    `json:"ownerId"`
	Notify      time.Time `json:"notify"`
}

type EventResp struct {
	Event Event `json:"event"`
}
