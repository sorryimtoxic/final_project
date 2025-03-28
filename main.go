package main

import (
	"log"
	"net/http"

	//"github.com/sorryimtoxic/go_final_project/api"
	 "go_final_project/api"
	//"github.com/sorryimtoxic/go_final_project/db"
	 "go_final_project/db"
)

func main() {
	const port = ":7540"
	const webDir = "./web"

	// Инициализирование БД
	database, err := db.InitDB(false)
	if err != nil {
		log.Fatalf("Ошибка при инициализации БД: %v", err)
	}
	defer database.Close()

	log.Println("БД успешно инициализирована.")

	// Настраиваем файловый сервер
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Обработчики для API
	http.HandleFunc("/api/nextdate", api.NextDateHandler)
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		api.GetTasksHandler(w, r, database) // Получение задач
	})
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.AddTaskHandler(w, r, database) // Добавление задач
		} else if r.Method == http.MethodDelete {
			api.DeleteTask(w, r) // Удаление задачи
		} else {
			api.TaskHandler(w, r) // Получение и обновление задач
		}
	})
	http.HandleFunc("/api/task/", api.TaskHandler)
	http.HandleFunc("/api/task/done", api.MarkTaskDone) // Регистрация обработчика для отметки о выполнении задачи

	log.Printf("Сервер запущен на http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
