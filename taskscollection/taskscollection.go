package taskscollection

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"os"
	"log"
        "reflect"
	"strconv"
	"time"
	"errors"
	"github.com/olekukonko/tablewriter"
	"github.com/dlt/selfcontrol/task"
	"github.com/deckarep/gosx-notifier"
)

var (
        tasksFilepath 		  = "/Users/dlt/golang/src/github.com/dlt/selfcontrol/tasks.json"
        ErrNoSuchTask             = errors.New("no such task")
	timers                    = make(map[int]taskTimer)
)

type Task task.Task

type taskTimer struct {
	Timer     *time.Timer
	StartedAt time.Time
	Task      *Task
}

type TasksCollection struct {
        Tasks []Task
        Outdated bool
}

func (tc *TasksCollection) Append(task Task)  {
        tc.Tasks = append(tc.Tasks, task)
}

func (tc *TasksCollection) GetTasks() []Task {
	if isDatabaseFileEmpty() {
		return make([]Task, 0)
	}
	if tc.Outdated {
                tc.Load()
	}
	return tc.Tasks
}

func (tc *TasksCollection) Find(id int) (*Task, error) {
	for _, t := range tc.GetTasks() {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, ErrNoSuchTask
}

func (tc *TasksCollection) StartTimerForTask(taskID int, d time.Duration) {
	task, err := tc.Find(taskID)
        if err != nil {
                panic(err)

        }
	timer := time.NewTimer(d)
	timers[task.ID] = taskTimer{
		Task:      task,
		Timer:     timer,
		StartedAt: time.Now(),
	}
	message := "Timer for '" + task.Name + "' finished!"
	go func() {
		<-timer.C
		pushNotification(message)
	}()
}

func (tc *TasksCollection) UpdateStatus(taskID int, status string) (bool, error) {
	task, err := tc.Find(taskID)
        if err != nil {
                return false, ErrNoSuchTask
        }
	task.Status = status
	tc.persist()
	return true, nil
}

func (tc *TasksCollection) Remove(id int) error {
        var newTasks []Task
        var noSuchTask = true
	for _, t := range tc.GetTasks() {
		if t.ID != id {
			newTasks = append(newTasks, t)
		} else {
                        noSuchTask = false
                }
	}
        if noSuchTask {
                return ErrNoSuchTask
        }
	tc.Tasks = newTasks
        tc.Outdated = true
        return nil
}

func (tc *TasksCollection) Load() {
        if tc.Outdated {
                tc.persist()
        }
	raw, err := ioutil.ReadFile(tasksFilepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(raw, &tc.Tasks)

        keys := reflect.ValueOf(timers).MapKeys()
        for k := range keys {
                task, err := tc.Find(k)
                if err == nil {
                        taskTimer := timers[k]
                        taskTimer.Task = task
                }
        }
}

func (tc *TasksCollection) maxID() int {
	id := 0

	for _, t := range tc.GetTasks() {
		if t.ID > id {
			id = t.ID
		}
	}
	return id
}

func (tc *TasksCollection) Create(name string) {
	task := Task{
		ID:                 0,
		Name:               name,
		Status:             "TODO",
		TotalTimeInSeconds: 0,
	}
	task.ID = tc.maxID() + 1
        tc.Append(task)
	tc.persist()
}

func (tc *TasksCollection) persist() {
        keys := reflect.ValueOf(timers).MapKeys()
        for k := range keys {
                t := timers[k]
                et, err := strconv.ParseFloat(ellapsedTime(*t.Task), 64)
                if err != nil {
                        panic(err)
                }
                t.Task.TotalTimeInSeconds = t.Task.TotalTimeInSeconds + et
        }
	tc.writeToDatabaseFile(toJSON(tc.Tasks))
}

func (tc *TasksCollection) Print()  {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "name", "status", "time", "ellapsed time"})

	for _, row := range tc.createRows() {
		table.Append(row)
	}
	table.Render()
}

func (tc *TasksCollection) writeToDatabaseFile(jsonString []byte) {
	f, err := os.Create(tasksFilepath)
	defer f.Close()

	if err != nil {
		fmt.Println("error:", err)
	}

	_, err = f.Write(jsonString)
	if err != nil {
		fmt.Println("error:", err)
	}
        tc.Outdated = false
}

func toJSON(tasks []Task) []byte {
	j, err := json.Marshal(tasks)
	if err != nil {
		fmt.Println("error:", err)
	}
	return j
}
func isDatabaseFileEmpty() bool {
	file, err := os.Open(tasksFilepath)

	if err != nil {
		log.Fatal(err)
	}
	f, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return f.Size() == 0
}

func ellapsedTime(t Task) string {
	timer := timers[t.ID]
	seconds := time.Now().Sub(timer.StartedAt).Seconds()
	if t.Status != "DOING" {
		return "-"
	}
	return strconv.FormatFloat(seconds, 'E', -1, 32)
}

func pushNotification(message string) {
	gosxnotifier.NewNotification(message).Push()
}


func (tc *TasksCollection) createRows() [][]string {
	var rows = [][]string{}

	for _, t := range tc.GetTasks() {
		row := []string{
			strconv.Itoa(t.ID),
			t.Name,
			t.Status,
			strconv.FormatFloat(t.TotalTimeInSeconds, 'E', -1, 64),
			ellapsedTime(t),
		}
		rows = append(rows, row)
	}

	return rows
}
