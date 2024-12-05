package app

import (
	"fmt"
	"net/http"

	"github.com/Fabriciuos/go_final_project_todolist/iternal/database"
	handlers "github.com/Fabriciuos/go_final_project_todolist/iternal/transport/rest"
	"github.com/go-chi/chi"
)

func Run() {
	r := chi.NewRouter()
	_, err := database.CreateDB()
	if err != nil {
		panic(err)
	}
	fmt.Println("Запускаем сервер на порте 7540")

	r.Handle("/*", http.FileServer(http.Dir("./web")))
	r.HandleFunc("/api/task/done", handlers.DoneTask)
	r.HandleFunc("/api/task", handlers.Task)
	r.HandleFunc("/api/nextdate", handlers.NextDeadLine)
	r.HandleFunc("/api/tasks", handlers.GetTasks)

	err = http.ListenAndServe(":7540", r)
	if err != nil {
		panic(err)
	}
}
