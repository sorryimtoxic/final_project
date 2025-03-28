package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AddTaskHandler обрабатывает POST-запросы для добавления задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Printf("Получен запрос: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Метод не разрешен"}`, http.StatusMethodNotAllowed)
		return
	}

	var task Task
	// Декодирование JSON из запроса
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error": "Ошибка десериализации JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if task.Title == "" {
		http.Error(w, `{"error": "Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// Проверка поля repeat
	if task.Repeat != "" && !isValidRepeat(task.Repeat) {
		http.Error(w, `{"error": "Недопустимое значение для поля repeat"}`, http.StatusBadRequest)
		return
	}

	// Проверка формата даты
	now := time.Now().Truncate(24 * time.Hour)
	var taskDate time.Time
	var err error

	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	taskDate, err = time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error": "Дата представлена в неправильном формате"}`, http.StatusBadRequest)
		return
	}

	if taskDate.Before(now) {
		taskDate = now
	} else if taskDate.Equal(now) {
		taskDate = now
	}

	// Вовзращает ошибку при указании еженедельного или ежемесячного правила повторения
	if task.Repeat != "" && (strings.HasPrefix(task.Repeat, "w") || strings.HasPrefix(task.Repeat, "m")) {
		http.Error(w, `{"error": "Неподдерживаемое правило повторения"}`, http.StatusBadRequest)
		return
	}

	// Устанавливаем task.Date на основе taskDate
	task.Date = taskDate.Format("20060102")

	// Добавление задачи в БД
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error": "Ошибка при добавлении задачи"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error": "Ошибка при получении идентификатора задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	response := map[string]interface{}{"id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "Ошибка при формировании ответа"}`, http.StatusInternalServerError)
		return
	}
}

// NextDate вычисляет следующую дату на основе правила повторения
func NextDate(now time.Time, date string, repeat string) (string, error) {
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("некорректная дата")
	}

	if repeat == "" {
		return "", errors.New("правило повторения не указано")
	}

	switch {
	case repeat == "y":
		// Ежегодное повторение
		nextDate := startDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil

	case strings.HasPrefix(repeat, "d "):
		// Проверка формата d <число>
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("неверный формат правила d")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("недопустимое количество дней")
		}

		nextDate := startDate.AddDate(0, 0, days)

		if nextDate.Equal(now) || nextDate.After(now) {
			return nextDate.Format("20060102"), nil
		}

		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}

		if nextDate.After(now) {
			return nextDate.Format("20060102"), nil
		}

		return "", errors.New("нет следующей даты")
	default:
		return "", errors.New("неподдерживаемый формат")
	}
}

// NextDateHandler обрабатывает запросы для получения следующей даты
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "некорректная дата now", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, nextDate)
}
