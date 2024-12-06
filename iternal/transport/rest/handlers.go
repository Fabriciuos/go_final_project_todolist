package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Fabriciuos/go_final_project_todolist/iternal/database"
	"github.com/Fabriciuos/go_final_project_todolist/iternal/nextdate"
	_ "github.com/mattn/go-sqlite3"
)

var (
	TimeFormat string = nextdate.TimeFormat
)

type TaskService struct {
	service database.TaskStorage
}

func NewTaskService(store database.TaskStorage) TaskService {
	return TaskService{service: store}
}

func (t TaskService) Task(w http.ResponseWriter, r *http.Request) {
	var task nextdate.Task
	var buf bytes.Buffer
	var date time.Time

	if r.Method == http.MethodGet {
		t.service.GetAllTasks()
		return
	} else if r.Method == http.MethodDelete {
		t.DeleteTask(w, r)
		return
	}

	now, _ := time.Parse(TimeFormat, time.Now().Format(TimeFormat))

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		callError("нет заголовка", w)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format(TimeFormat)
		date, _ = time.Parse(TimeFormat, time.Now().Format(TimeFormat))
	} else {
		date, err = time.Parse(TimeFormat, task.Date)
		if err != nil {
			callError("неверный формат даты", w)
			return
		}
	}

	if now.After(date) {
		if task.Repeat == "" {
			task.Date = time.Now().Format(TimeFormat)
		} else {
			task.Date, err = nextdate.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				callError("неверный формат", w)
				return
			}
		}
	}
	if r.Method == http.MethodPut {
		t.EditTask(w, r, task)
		return
	}

	id, err := t.service.PutTaskInDB(task)
	if err != nil {
		callError("ошибка с базой данных", w)
		return
	}

	resp, err := json.Marshal(map[string]string{"id": strconv.Itoa(int(id))})
	if err != nil {
		callError("не получилось создать напоминание", w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)

}

func (t TaskService) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := []nextdate.Task{}

	count, err := t.service.GetCountOfTasks()
	if err != nil {
		callError("ошибка с базой данных", w)
		return
	}

	if count > 0 {
		tasks, err = t.service.GetAllTasks()
		if err != nil {
			callError("ошибка с базой данных", w)
			return
		}
	}
	resp, err := json.Marshal(map[string]interface{}{
		"tasks": tasks,
	})
	if err != nil {
		callError("ошибка сериализации данных", w)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)

}

func NextDeadLine(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse(TimeFormat, r.URL.Query().Get("now"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	deadline, err := nextdate.NextDate(now, date, repeat)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	_, _ = w.Write([]byte(deadline))

}

func (t TaskService) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	var task nextdate.Task

	id := r.FormValue("id")
	row, err := t.service.GetTask(id)
	if err != nil {
		callError("ошибка с базой данных", w)
		return
	}

	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		callError("Задача не найдена", w)
		return
	}

	resp, err := json.Marshal(map[string]string{
		"id":      task.ID,
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	})
	if err != nil {
		callError("ошибка сериализации данных", w)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (t TaskService) EditTask(w http.ResponseWriter, h *http.Request, task nextdate.Task) {
	var checkerrortask nextdate.Task
	row, _ := t.service.GetTask(task.ID)
	err := row.Scan(&checkerrortask.ID, &checkerrortask.Date, &checkerrortask.Title, &checkerrortask.Comment, &checkerrortask.Repeat)
	if err != nil {
		callError("задача не найдена", w)
		return

	}
	err = t.service.EditTask(task)
	if err != nil {
		callError("ошибка подключения к базе данных", w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, _ = w.Write([]byte("{}"))

}

func (t TaskService) DoneTask(w http.ResponseWriter, r *http.Request) {
	var task nextdate.Task

	now, _ := time.Parse(TimeFormat, time.Now().Format(TimeFormat))

	id := r.FormValue("id")
	row, err := t.service.GetTask(id)
	if err != nil {
		callError("ошибка с базой данных", w)
		return
	}

	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		callError("Задача не найдена", w)
		return
	}

	if task.Repeat == "" {
		err = t.service.DeleteTask(task.ID)
		if err != nil {
			callError("не получилоось отметить задачу выполненной", w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte("{}"))
		return
	} else {
		task.Date, err = nextdate.NextDate(now, task.Date, task.Repeat)
	}
	if err != nil {
		callError("не получилось найти следующую дату", w)
		return
	}
	err = t.service.EditTask(task)
	if err != nil {
		callError("не получилось обновить дату в задаче", w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, _ = w.Write([]byte("{}"))
}

func (t TaskService) DeleteTask(w http.ResponseWriter, r *http.Request) {
	var task nextdate.Task

	id := r.FormValue("id")
	row, err := t.service.GetTask(id)
	if err != nil {
		callError("ошибка с базой данных", w)
		return
	}

	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		callError("Задача не найдена", w)
		return
	}

	err = t.service.DeleteTask(task.ID)
	if err != nil {
		callError("не получилось удалить задачу", w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, _ = w.Write([]byte("{}"))
}

func callError(txt string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"error": txt})
}
