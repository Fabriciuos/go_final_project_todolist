package database

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/Fabriciuos/go_final_project_todolist/iternal/nextdate"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFile = "project.db"
	limit  = "50"
)

type TaskStorage struct {
	db *sql.DB
}

func NewTaskStorage(db *sql.DB) TaskStorage {
	return TaskStorage{db: db}
}

func CreateDB() (*sql.DB, error) {
	//appPath, err := os.Executable()
	//if err != nil {
	//log.Fatal(err)
	//}
	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		return nil, err
	}
	defer db.Close()

	if install {
		query := ` 
		CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT CHECK(LENGTH(repeat) <= 128)
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);`
		_, err = db.Exec(query)
		if err != nil {
			return nil, err
		}
		log.Println("База данных создана")
	}
	return db, nil
}

func (t TaskStorage) PutTaskInDB(task nextdate.Task) (int64, error) {
	res, err := t.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (t TaskStorage) GetCountOfTasks() (int, error) {
	var count int64

	row := t.db.QueryRow("SELECT count(*) FROM scheduler")
	_ = row.Scan(&count)

	return int(count), nil
}

func (t TaskStorage) GetAllTasks() ([]nextdate.Task, error) {
	var tasks []nextdate.Task

	rows, err := t.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var task nextdate.Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, errors.New("ошибка с базой данных")
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (t TaskStorage) GetTask(id string) (*sql.Row, error) {
	row := t.db.QueryRow("SELECT * FROM scheduler WHERE id=?", id)

	return row, nil
}

func (t TaskStorage) EditTask(task nextdate.Task) error {
	_, err := t.db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	return nil
}

func (t TaskStorage) DeleteTask(id string) error {
	_, err := t.db.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		return err
	}

	return nil
}
