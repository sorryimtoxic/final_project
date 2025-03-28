package api

import (
	"net/http"
	"strconv"
	"time"

	"go_final_project/db"
)

// Обработчик для отметки о выполнении задачи
func MarkTaskDone(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// Получение задачи из БД
	task, err := db.GetTaskByID(strconv.Itoa(id))
	if err != nil {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	if task.Repeat != "" {
		// Рассчет следующей даты
		currentDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, `{"error":"Неверный формат даты"}`, http.StatusBadRequest)
			return
		}

		nextDate, err := NextDate(currentDate, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Ошибка при расчете следующей даты"}`, http.StatusInternalServerError)
			return
		}

		task.Date = nextDate 
	} else {
		err = db.DeleteTaskByID(id)
		if err != nil {
			http.Error(w, `{"error":"Ошибка при удалении задачи"}`, http.StatusInternalServerError)
			return
		}
		w.Write([]byte("{}"))
		return
	}

	err = db.UpdateTask(task)
	if err != nil {
		http.Error(w, `{"error":"Ошибка при обновлении задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}

// Удаление задачи
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Неверный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	err = db.DeleteTaskByID(id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка при удалении задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}
