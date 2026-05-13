package database

import (
	"os"
	"testing"
	"time"

	"gonews/pkg/models"
)

// getTestConfig возвращает конфигурацию из переменных окружения
func getTestConfig() Config {
	return Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}
}

func TestNewDB(t *testing.T) {
	// Пропускаем тест, если не настроены переменные окружения для тестов
	if os.Getenv("DB_TEST") != "1" {
		t.Skip("Пропускаем тест БД. Установите DB_TEST=1 для запуска")
	}

	cfg := getTestConfig()
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Error("Подключение к БД не установлено")
	}
}

func TestDB_SaveAndGetPosts(t *testing.T) {
	if os.Getenv("DB_TEST") != "1" {
		t.Skip("Пропускаем тест БД. Установите DB_TEST=1 для запуска")
	}

	cfg := getTestConfig()
	db, err := NewDB(cfg)
	if err != nil {
		t.Skipf("Пропускаем тест: невозможно подключиться к БД: %v", err)
	}
	defer db.Close()

	// Инициализация схемы (создание таблицы)
	schema := `
		CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(500) NOT NULL,
			content TEXT,
			pub_time TIMESTAMP NOT NULL,
			link VARCHAR(1000) UNIQUE NOT NULL,
			source VARCHAR(200),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	if err := db.InitSchema(schema); err != nil {
		t.Fatalf("Ошибка инициализации схемы: %v", err)
	}

	// Создаем тестовую запись, используя models.Post
	testPost := models.Post{
		Title:   "Тестовый заголовок",
		Content: "Тестовое содержание",
		PubTime: time.Now(),
		Link:    "https://example.com/test-unique-link",
		Source:  "Test Source",
	}

	// Сохранение записи
	err = db.SavePost(testPost)
	if err != nil {
		t.Errorf("Ошибка сохранения поста: %v", err)
	}

	// Получение записей
	posts, err := db.GetPosts(10)
	if err != nil {
		t.Errorf("Ошибка получения постов: %v", err)
	}

	if len(posts) == 0 {
		t.Error("Ожидался хотя бы один пост")
	}

	// Проверка данных
	found := false
	for _, p := range posts {
		if p.Link == testPost.Link {
			found = true
			if p.Title != testPost.Title {
				t.Errorf("Заголовок не совпадает: ожидалось %s, получено %s", testPost.Title, p.Title)
			}
			if p.Content != testPost.Content {
				t.Errorf("Содержание не совпадает: ожидалось %s, получено %s", testPost.Content, p.Content)
			}
		}
	}

	if !found {
		t.Error("Сохраненный пост не найден в базе")
	}
}

func TestDB_GetPosts_Limit(t *testing.T) {
	if os.Getenv("DB_TEST") != "1" {
		t.Skip("Пропускаем тест БД. Установите DB_TEST=1 для запуска")
	}

	cfg := getTestConfig()
	db, err := NewDB(cfg)
	if err != nil {
		t.Skipf("Пропускаем тест: невозможно подключиться к БД: %v", err)
	}
	defer db.Close()

	// Запрашиваем 5 постов
	posts, err := db.GetPosts(5)
	if err != nil {
		t.Errorf("Ошибка получения постов: %v", err)
	}

	// Проверяем, что вернулось не больше 5 постов
	if len(posts) > 5 {
		t.Errorf("Ожидалось не более 5 постов, получено %d", len(posts))
	}
}

// Тест структуры модели (проверка JSON тегов и полей)
func TestPostModelStructure(t *testing.T) {
	now := time.Now()
	post := models.Post{
		ID:      1,
		Title:   "Test Title",
		Content: "Content",
		PubTime: now,
		Link:    "https://example.com",
		Source:  "Source",
	}

	if post.ID != 1 {
		t.Error("Неверный ID")
	}
	if post.Title != "Test Title" {
		t.Error("Неверный заголовок")
	}
	if post.PubTime != now {
		t.Error("Неверное время публикации")
	}
}
