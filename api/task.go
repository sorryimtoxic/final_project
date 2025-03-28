package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Task представляет структуру задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var tasks []Task
	limit := 10

	// Получаем все задачи из БД
	query := "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?;"
	rows, err := db.Query(query, limit)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %v\n", err)
		http.Error(w, `{"error": "Ошибка при получении задач: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		var dateStr string

		if err := rows.Scan(&task.ID, &dateStr, &task.Title, &task.Comment, &task.Repeat); err != nil {
			log.Printf("Ошибка при считывании задач: %v\n", err)
			http.Error(w, `{"error": "Ошибка при считывании задач: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		// Проверка формата даты
		now := time.Now().Truncate(24 * time.Hour)
		if dateStr == "" {
			task.Date = now.Format("20060102")
		} else {
			date, err := time.Parse("20060102", dateStr)
			if err != nil {
				log.Printf("Ошибка при парсинге даты: %v\n", err)
				http.Error(w, `{"error": "Неверный формат даты"}`, http.StatusBadRequest)
				return
			}
			task.Date = date.Format("20060102")
		}
		tasks = append(tasks, task)
	}

	if len(tasks) < 1 {
		tasks = make([]Task, 0)
	}

	// Формирование ответа
	response := map[string]interface{}{"tasks": tasks}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка при формировании ответа: %v\n", err)
		http.Error(w, `{"error": "Ошибка при формировании ответа: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
}
