package tasks

import (
	"encoding/json"
	"errors"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/deckarep/gosx-notifier"
	"github.com/olekukonko/tablewriter"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	myDBDir        = "/Users/dlt/.selfcontrol"
	collectionName = "Tasks"
)

var (
	errDuplicateTimerForTask = errors.New("task already has a running timer")
	tasksCollection          *db.Col
	timers                   = make(map[int][]*taskTimer)
)

type task map[string]interface{}

type taskTimer struct {
	Timer      *time.Timer
	StartedAt  time.Time
	FinishedAt time.Time
}

func (tt *taskTimer) notify(message string) {
	pushNotification(message)
	tt.FinishedAt = time.Now()
}

func init() {
	tasksDB, err := db.OpenDB(myDBDir)
	if err != nil {
		panic(err)
	}
	tasksDB.Create(collectionName)
	tasksCollection = tasksDB.Use(collectionName)
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
	table.SetHeader([]string{"ID", "name", "status"})
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
	taskTimer := &taskTimer{timer, time.Now(), time.Time{}}

	timers[taskID] = append(timers[taskID], taskTimer)

	go func() {
		<-timer.C
		taskTimer.notify(message)
	}()

	return true, nil
}

func hasRunningTimer(taskID int) bool {
	zeroedTime := time.Time{}
	for _, timer := range timers[taskID] {
		if timer.FinishedAt == zeroedTime {
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
		}
		rows = append(rows, row)
		return true
	})
	return rows
}

func coloredStatus(status string) string  {
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
