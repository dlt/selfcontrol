package tasks

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/deckarep/gosx-notifier"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
	"time"
)

const myDBDir = "/Users/dlt/.selfcontrol"

var (
	errDuplicateTimerForTask = errors.New("task already has a running timer")
	tasksCollection          *db.Col
	timersCollection         *db.Col
	DB                       *db.DB
	timers                   map[int][]*taskTimer
)

type task map[string]interface{}

type taskTimer struct {
	TaskID       int
	Timer        *time.Timer
	StartedAt    time.Time
	FinishedAt   time.Time
	Message      string
	Fired        bool
	Persisted    bool
}

func (tt *taskTimer) trigger() {
	pushNotification(tt.Message)
	tt.FinishedAt = time.Now()
	tt.Fired = true
	tt.Persisted = true
	tt.save()
}

func (tt *taskTimer) ellapsedTime() time.Duration {
	var finishedAt time.Time
	if tt.unfinished() {
		finishedAt = time.Now()
	} else {
		finishedAt = tt.FinishedAt
	}
	return finishedAt.Sub(tt.StartedAt)
}

func (timer *taskTimer) unfinished() bool {
	return !timer.Fired
}

func (timer *taskTimer) save() {
	timersDoc := map[string]interface{}{
		"TaskID":       string(timer.TaskID),
		"Fired":        timer.Fired,
		"Message": 	timer.Message,
		"StartedAt":    timer.StartedAt,
		"FinishedAt":   timer.FinishedAt,
	}
	_, err := timersCollection.Insert(timersDoc)
	if err != nil {
		panic(err)
	}
}

func Save()  {
	for _, taskTimers := range timers {
		for _, timer := range taskTimers {
			if !timer.Persisted {
				timer.save()
			}
		}
	}
}

func init() {
	d, err := db.OpenDB(myDBDir)
	DB = d
	if err != nil {
		panic(err)
	}
	DB.Create("Tasks")
	DB.Create("Timers")
	tasksCollection = DB.Use("Tasks")
	timersCollection = DB.Use("Timers")
	timers = make(map[int][]*taskTimer)
	loadTimers()
}

func loadTimers() {
	timersCollection.ForEachDoc(func(id int, raw []byte) (willMoveOn bool) {
		var doc map[string]interface{}
		json.Unmarshal(raw, &doc)
		startedAt, err := time.Parse(time.RFC3339, doc["StartedAt"].(string))
		if err != nil {
			panic(err)
		}
		finishedAt, err := time.Parse(time.RFC3339, doc["FinishedAt"].(string))
		if err != nil {
			panic(err)
		}
		timer := &taskTimer{
			TaskID:       id,
			Fired:        doc["Fired"].(bool),
			StartedAt:    startedAt,
			FinishedAt:   finishedAt,
		}
		timers[id] = append(timers[id], timer)
		return true
	})
}

// Create a new task given its name
func Create(name string) {
	task := map[string]interface{}{
		"name":   name,
		"status": "TODO",
	}
	_, err := tasksCollection.Insert(task)
	if err != nil {
		panic(err)
	}
}

// Print all tasks in a ASCII table
func Print() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "name", "status", "total time"})
	table.SetAutoFormatHeaders(false)
	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()
}

// Delete a task with given id
func Delete(id int) {
	if err := tasksCollection.Delete(id); err != nil {
		panic(err)
	}
}

// UpdateFields updates the tasks attributes
func UpdateFields(taskID int, fieldValuePairs []string) (bool, error) {
	task, err := tasksCollection.Read(taskID)
	if err != nil {
		return false, err
	}
	for _, pair := range fieldValuePairs {
		arr := strings.Split(pair, ":")
		field, value := arr[0], arr[1]
		if field == "status" {
			value = strings.ToUpper(value)
		}
		task[field] = value
	}
	if err = tasksCollection.Update(taskID, task); err != nil {
		return false, err
	}
	return true, nil
}

// AddTimerForTask adds a timer for a given task id and duration
func AddTimerForTask(taskID int, d time.Duration) (bool, error) {
	task, err := tasksCollection.Read(taskID)
	if err != nil {
		panic(err)
	}

	if hasRunningTimer(taskID) {
		return false, errDuplicateTimerForTask
	}

	name := task["name"].(string)
	timer := time.NewTimer(d)
	message := "Timer for '" + name + "' finished!"
	tTimer := &taskTimer{
		TaskID:     taskID,
		Timer:      timer,
		Message:    message,
		StartedAt:  time.Now(),
		FinishedAt: time.Time{},
	}

	if timers[taskID] == nil {
		timers[taskID] = make([]*taskTimer, 0)
	}

	timers[taskID] = append(timers[taskID], tTimer)

	go func() {
		<-timer.C
		tTimer.trigger()
		Print()
	}()
	return true, nil
}

func hasRunningTimer(taskID int) bool {
	for _, timer := range timers[taskID] {
		if timer.unfinished() {
			return true
		}
	}
	return false
}

func pushNotification(message string) {
	gosxnotifier.NewNotification(message).Push()
}

func createRows() [][]string {
	var rows = [][]string{}
	tasksCollection.ForEachDoc(func(id int, raw []byte) (willMoveOn bool) {
		var doc task
		json.Unmarshal(raw, &doc)
		name := doc["name"].(string)
		status := doc["status"].(string)
		row := []string{
			strconv.Itoa(id),
			name,
			coloredStatus(status),
			totalRunningTime(id).String(),
		}
		rows = append(rows, row)
		return true
	})
	return rows
}

func totalRunningTime(taskID int) time.Duration {
	loadTimers()
	total := time.Duration(0)
	for _, timer := range timers[taskID] {
		total += timer.ellapsedTime()
	}
	return total
}

func coloredStatus(status string) string {
	switch status {
	case "TODO":
		return color.CyanString(status)
	case "DOING":
		return color.YellowString(status)
	case "DONE":
		return color.GreenString(status)
	}
	return ""
}
