package taskscollection

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"github.com/HouzuoGuo/tiedot/db"
	_ "github.com/HouzuoGuo/tiedot/dberr"
	_ "github.com/deckarep/gosx-notifier"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	myDBDir         = "/Users/dlt/.selfcontrol"
	collectionName  = "Tasks"
	ErrNoSuchTask   = errors.New("no such task")
	tasksCollection *db.Col
	timers          = make(map[int]taskTimer)
)

type Task map[string]interface{}

type taskTimer struct {
	Timer     *time.Timer
	StartedAt time.Time
	Task      *Task
}

func init() {
	tasksDB, err := db.OpenDB(myDBDir)
	if err != nil {
		panic(err)
	}
	tasksDB.Create(collectionName)
	tasksCollection = tasksDB.Use(collectionName)
}

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

func Print() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "name", "status"})
	table.SetAutoFormatHeaders(false)
	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()
}

func Delete(id int) {
	if err := tasksCollection.Delete(id); err != nil {
		panic(err)
	}
}

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

func StartTimerForTask(taskID int, d time.Duration) {
	task, err := tasksCollection.Read(taskID)
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

func pushNotification(message string) {
	gosxnotifier.NewNotification(message).Push()
}

func createRows() [][]string {
	var rows = [][]string{}
	tasksCollection.ForEachDoc(func(id int, raw []byte) (willMoveOn bool) {
		var doc Task
		json.Unmarshal(raw, &doc)
		name := doc["name"].(string)
		status := doc["status"].(string)
		row := []string{
			strconv.Itoa(id),
			name,
			status,
		}
		rows = append(rows, row)
		return true
	})
	return rows
}

func ellapsedTime(t Task) string {
	timer := timers[t.ID]
	seconds := time.Now().Sub(timer.StartedAt).Seconds()
	if t.Status != "DOING" {
		return "-"
	}
	return strconv.FormatFloat(seconds, 'E', -1, 32)
}
