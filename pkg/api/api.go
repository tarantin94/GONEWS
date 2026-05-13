// Package api предоставляет HTTP API для агрегатора новостей
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gonews/pkg/database"

	"github.com/gorilla/mux"
)

// API представляет HTTP API сервер с маршрутизацией
type API struct {
	router *mux.Router
	db     *database.DB
}

// NewAPI создает новый экземпляр API сервера с настроенными маршрутами
func NewAPI(db *database.DB) *API {
	api := &API{
		router: mux.NewRouter(),
		db:     db,
	}
	api.setupRoutes()
	return api
}

// setupRoutes настраивает маршруты для API и раздачи статики
func (a *API) setupRoutes() {
	// Эндпоинт для получения новостей: /news/{n}
	a.router.HandleFunc("/news/{n}", a.getNews).Methods("GET", "OPTIONS")
	// Раздача статических файлов веб-интерфейса
	a.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp"))))
}

// getNews обработчик HTTP-запроса для получения последних новостей
func (a *API) getNews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nStr := vars["n"]

	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		http.Error(w, "Неверное количество новостей", http.StatusBadRequest)
		return
	}

	// Ограничиваем максимальное количество новостей
	if n > 100 {
		n = 100
	}

	posts, err := a.db.GetPosts(n)
	if err != nil {
		http.Error(w, "Ошибка получения новостей", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(posts)
}

// GetRouter возвращает настроенный маршрутизатор для использования в http.Server
func (a *API) GetRouter() *mux.Router {
	return a.router
}

// ServeHTTP реализует интерфейс http.Handler
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
