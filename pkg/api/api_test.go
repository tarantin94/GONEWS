package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gonews/pkg/database"
)

func TestAPI_GetNews(t *testing.T) {
	if os.Getenv("DB_TEST") != "1" {
		t.Skip("Пропускаем тест API. Установите DB_TEST=1 для запуска")
	}

	cfg := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		t.Skipf("Пропускаем тест: невозможно подключиться к БД: %v", err)
	}
	defer db.Close()

	api := NewAPI(db)

	// Тест запроса
	req := httptest.NewRequest("GET", "/news/10", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получен %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Ожидался Content-Type application/json, получен %s", contentType)
	}
}

func TestAPI_GetNews_InvalidNumber(t *testing.T) {
	cfg := database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "gonews",
		SSLMode:  "disable",
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		t.Skip("Пропускаем тест: невозможно подключиться к БД")
	}
	defer db.Close()

	api := NewAPI(db)

	// Тест с невалидным числом
	req := httptest.NewRequest("GET", "/news/abc", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус 400, получен %d", w.Code)
	}
}

func TestAPI_Router(t *testing.T) {
	cfg := database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "gonews",
		SSLMode:  "disable",
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		t.Skip("Пропускаем тест: невозможно подключиться к БД")
	}
	defer db.Close()

	api := NewAPI(db)
	router := api.GetRouter()

	if router == nil {
		t.Error("Маршрутизатор не должен быть nil")
	}
}
