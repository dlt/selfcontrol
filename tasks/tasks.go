package tasks

import (
	"errors"
	_ "fmt"
	"github.com/deckarep/gosx-notifier"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
 	"github.com/asdine/storm"
	"os"
	"strconv"
	_ "strings"
	"time"
)

var (
	errDuplicateTimerForTask = errors.New("task already has a running timer")
	ticker 			 <-chan time.Time
	DB			 *storm.DB
)

type task struct {
	ID int `storm:"id,increment"`
	Name string `storm:"unique"`
	Status string
}

type taskTimer struct {
	ID         int `storm:"id,increment"`
	TaskID     int
	StartedAt  time.Time
	ExpiresAt  time.Time
	Message    string
	Fired      bool
}

func (tt *taskTimer) trigger() {
	pushNotification(tt.Message)
	tt.Fired = true
	tt.update()
}

func (tt *taskTimer) ellapsedTime() time.Duration {
	var finishedAt time.Time
	if tt.unfinished() {
		finishedAt = time.Now()
	} else {
		finishedAt = tt.ExpiresAt
	}
	return finishedAt.Sub(tt.StartedAt)
}

func (timer *taskTimer) unfinished() bool {
	return !timer.Fired
}

func (timer *taskTimer) update() {
	err := DB.Update(&timer)
	if err != nil {
		panic(err)
	}
}

func init() {
	d, err := storm.Open("selfcontrol.db")
	DB = d
	if err != nil {
		panic(err)
	}
	startTimersLoop()
}

func startTimersLoop()  {
    ticker = time.Tick(1 * time.Second)
    go func ()  {
	    for _ = range ticker {
	    }
    }()
}

// Create a new task given its name
func Create(name string) {
	timer := task{
		Name: name,
		Status: "TODO",
	}
	err := DB.Save(&timer)
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
func Delete(ID int) {
	var t task
	err := DB.Find("ID", ID, &t)
	if err != nil {
		panic(err)
	}
	if err = DB.DeleteStruct(&t); err != nil {
		panic(err)
	}
}

// UpdateFields updates the tasks attributes
// func UpdateFields(taskID int, fieldValuePairs []string) (bool, error) {
// 	task, err := tasksCollection.Read(taskID)
// 	if err != nil {
// 		return false, err
// 	}
// 	for _, pair := range fieldValuePairs {
// 		arr := strings.Split(pair, ":")
// 		field, value := arr[0], arr[1]
// 		if field == "status" {
// 			value = strings.ToUpper(value)
// 		}
// 		task[field] = value
// 	}
// 	if err = tasksCollection.Update(taskID, task); err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }

// AddTimerForTask adds a timer for a given task id and duration
func AddTimerForTask(taskID int, d time.Duration) (bool, error) {
	var t task
	err := DB.Find("ID", taskID, &t)
	if err != nil {
		panic(err)
	}

//	if hasRunningTimer(taskID) {
//		return false, errDuplicateTimerForTask
//	}

	name := t.Name
	startedAt := time.Now()
	expiresAt := startedAt.Add(d)

	DB.Save(&taskTimer{
		TaskID:     taskID,
		Message:    "Timer for '" + name + "' finished!",
		StartedAt:  startedAt,
		ExpiresAt:  expiresAt,
	})
	return true, nil
}

// func hasRunningTimer(taskID int) bool {
// 	for _, timer := range timers[taskID] {
// 		if timer.unfinished() {
// 			return true
// 		}
// 	}
// 	return false
// }

func pushNotification(message string) {
	gosxnotifier.NewNotification(message).Push()
}

func createRows() [][]string {
	var rows = [][]string{}
	var tasks []task
	err := DB.All(&tasks)
	if err != nil {
		panic(err)
	}

	for _, tt := range tasks {
		row := []string{
			strconv.Itoa(tt.ID),
			tt.Name,
			coloredStatus(tt.Status),
			totalRunningTime(tt.ID).String(),
		}
		rows = append(rows, row)
	}
	return rows
}

func totalRunningTime(taskID int) time.Duration {
	total := time.Duration(0)
	var taskTimers []taskTimer
	err := DB.Find("ID", taskID, &taskTimers)
	if err != nil {
		panic(err)
	}
	for _, t := range taskTimers {
		total += t.ellapsedTime()
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
