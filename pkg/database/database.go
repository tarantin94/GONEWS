// Package database предоставляет функции для работы с базой данных PostgreSQL
package database

import (
	"database/sql"
	"fmt"

	"gonews/pkg/models"

	_ "github.com/lib/pq"
)

// DB обертка над sql.DB для удобной работы с базой данных
type DB struct {
	db *sql.DB
}

// Config конфигурация подключения к базе данных
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB создает новое подключение к базе данных по конфигурации
func NewDB(cfg Config) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к БД: %w", err)
	}

	return &DB{db: db}, nil
}

// Close закрывает подключение к базе данных
func (d *DB) Close() error {
	return d.db.Close()
}

// SavePost сохраняет одну публикацию в базе данных
// При конфликте по полю link (уникальный) запись пропускается
func (d *DB) SavePost(post models.Post) error {
	query := `
		INSERT INTO posts (title, content, pub_time, link, source)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (link) DO NOTHING`

	_, err := d.db.Exec(query, post.Title, post.Content, post.PubTime, post.Link, post.Source)
	return err
}

// SavePosts сохраняет несколько публикаций, возвращает количество успешно сохранённых
func (d *DB) SavePosts(posts []models.Post) (int, error) {
	saved := 0
	for _, post := range posts {
		if err := d.SavePost(post); err == nil {
			saved++
		}
	}
	return saved, nil
}

// GetPosts получает последние n публикаций из базы данных, отсортированные по дате
func (d *DB) GetPosts(n int) ([]models.Post, error) {
	query := `
		SELECT id, title, content, pub_time, link, source
		FROM posts
		ORDER BY pub_time DESC
		LIMIT $1`

	rows, err := d.db.Query(query, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link, &p.Source)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

// InitSchema инициализирует схему базы данных, выполняя переданный SQL-скрипт
func (d *DB) InitSchema(schemaSQL string) error {
	_, err := d.db.Exec(schemaSQL)
	return err
}

// GetPostByID получает одну новость по ID
func (d *DB) GetPostByID(id int) (models.Post, error) {
	query := `
		SELECT id, title, content, pub_time, link, source
		FROM posts
		WHERE id = $1`

	var post models.Post
	err := d.db.QueryRow(query, id).Scan(
		&post.ID, &post.Title, &post.Content,
		&post.PubTime, &post.Link, &post.Source,
	)
	return post, err
}

// GetPostsPaginated получает новости с пагинацией (без поиска)
func (d *DB) GetPostsPaginated(perPage, offset int) ([]models.Post, int, error) {
	// 1. Получаем общее количество новостей
	var total int
	err := d.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 2. Получаем нужную страницу
	query := `
		SELECT id, title, content, pub_time, link, source
		FROM posts
		ORDER BY pub_time DESC
		LIMIT $1 OFFSET $2`

	rows, err := d.db.Query(query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link, &p.Source); err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}

	return posts, total, rows.Err()
}

// GetPostsPaginatedWithSearch получает новости с поиском по названию и пагинацией
// Использует ILIKE для регистронезависимого поиска
func (d *DB) GetPostsPaginatedWithSearch(search string, perPage, offset int) ([]models.Post, int, error) {
	searchPattern := "%" + search + "%"

	// 1. Получаем общее количество новостей с учётом фильтра
	var total int
	err := d.db.QueryRow(
		"SELECT COUNT(*) FROM posts WHERE title ILIKE $1",
		searchPattern,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 2. Получаем нужную страницу с фильтром
	query := `
		SELECT id, title, content, pub_time, link, source
		FROM posts
		WHERE title ILIKE $1
		ORDER BY pub_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := d.db.Query(query, searchPattern, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link, &p.Source); err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}

	return posts, total, rows.Err()
}
