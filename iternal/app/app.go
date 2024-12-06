package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Fabriciuos/go_final_project_todolist/iternal/database"
	handlers "github.com/Fabriciuos/go_final_project_todolist/iternal/transport/rest"
	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
)

func Run() {
	r := chi.NewRouter()
	_, err := database.CreateDB()
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlite3", "project.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	store := database.NewTaskStorage(db)
	service := handlers.NewTaskService(store)
	fmt.Println("Запускаем сервер на порте 7540")

	r.Handle("/*", http.FileServer(http.Dir("./web")))
	r.HandleFunc("/api/task/done", service.DoneTask)
	r.HandleFunc("/api/task", service.Task)
	r.HandleFunc("/api/nextdate", handlers.NextDeadLine)
	r.HandleFunc("/api/tasks", service.GetTasks)

	err = http.ListenAndServe(":7540", r)
	if err != nil {
		panic(err)
	}
}
