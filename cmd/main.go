package main

import (
	"database/sql"

	"github.com/Fabriciuos/go_final_project_todolist/iternal/app"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sql.Open("sqlite3", "project")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	app.Run()
}
