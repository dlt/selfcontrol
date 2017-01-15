package tasks

import (
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/deckarep/gosx-notifier"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	errDuplicateTimerForTask = errors.New("task already has a running timer")
	ticker                   <-chan time.Time
	DB                       *storm.DB
)

type task struct {
	ID       int    `storm:"id,increment"`
	Name     string `storm:"unique"`
	Status   string
	Priority int
	Tags     []string
}

func (t *task) updateFieldValuePairs(fieldValuePairs []string) {
	for _, pair := range fieldValuePairs {
		arr := strings.Split(pair, ":")
		field, value := arr[0], arr[1]
		if field == "status" {
			value = strings.ToUpper(value)
		}
		t.updateField(field, value)
	}
}

func (t *task) updateField(field, value string) {
	switch field {
	case "name", "n":
		DB.Update(&task{ID: t.ID, Name: value})
	case "status", "st", "s":
		DB.Update(&task{ID: t.ID, Status: strings.ToUpper(value)})
	case "priority", "pri", "p":
		priority, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		DB.Update(&task{ID: t.ID, Priority: priority})
	case "tags", "t":
		tags := processTags(t.Tags, value)
		DB.Update(&task{ID: t.ID, Tags: tags})
	}
}

func processTags(oldTags []string, tags string) []string {
	if len(strings.Trim(tags, " ")) == 0 {
		return oldTags
	}
	newTags := make([]string, 0)
	if strings.Index(tags, ",") > 0 {
		newTags = append(newTags, strings.Split(tags, ",")...)
	} else {
		newTags = append(newTags, tags)
	}
	for _, tag := range newTags {
		if tag[0] == '-' {
			oldTags = removeTag(oldTags, tag[1:])
		} else {
			oldTags = addTag(oldTags, tag)
		}
	}

	return oldTags
}

func addTag(tags []string, tag string) []string {
	for _, t := range tags {
		if t == tag {
			return tags
		}
	}
	return append(tags, tag)
}

func removeTag(tags []string, tag string) []string {
	newTags := make([]string, 0)
	for _, t := range tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	return newTags
}

type taskTimer struct {
	ID        int `storm:"id,increment"`
	TaskID    int
	StartedAt time.Time
	ExpiresAt time.Time
	Message   string
	Fired     bool
}

func (tt *taskTimer) trigger() {
	tt.pushNotification()
	err := DB.Update(&taskTimer{ID: tt.ID, Fired: true})
	if err != nil {
		panic(err)
	}
}

func (tt *taskTimer) isExpired() bool {
	return tt.ExpiresAt.Before(time.Now())
}

func (tt *taskTimer) ellapsedTime() time.Duration {
	var finishedAt time.Time
	if !tt.Fired {
		finishedAt = time.Now()
	} else {
		finishedAt = tt.ExpiresAt
	}
	return finishedAt.Sub(tt.StartedAt)
}

func Init(dbfile string) {
	d, err := storm.Open(dbfile)
	DB = d
	if err != nil {
		panic(err)
	}
	startTimersLoop()
}

func startTimersLoop() {
	ticker = time.Tick(1 * time.Second)
	go func() {
		for _ = range ticker {
			var taskTimers []taskTimer
			_ = DB.Find("Fired", false, &taskTimers)
			for _, tt := range taskTimers {
				if tt.isExpired() {
					tt.trigger()
				}
			}
		}
	}()
}

// Add a new task given its name
func Add(name string, fieldValuePairs []string) {
	t := &task{Status: "TODO"}
	err := DB.Save(t)
	if err != nil {
		panic(err)
	}
	fieldValuePairs = append(fieldValuePairs, "n:"+name)
	t.updateFieldValuePairs(fieldValuePairs)
}

// Print all tasks in a ASCII table
func Print() {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"ID", "Name", "Status", "Priority", "Tags", "Total time"}
	coloredHeader := make([]string, 0)
	for _, h := range header {
		coloredHeader = append(coloredHeader, color.YellowString(h))
	}
	table.SetHeader(coloredHeader)
	table.SetAutoFormatHeaders(false)
	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()
}

// Delete a task with given id
func Delete(ID int) {
	var t task
	err := DB.One("ID", ID, &t)
	if err != nil {
		panic(err)
	}
	if err = DB.DeleteStruct(&t); err != nil {
		panic(err)
	}
}

//UpdateFields updates the tasks attributes
func UpdateFields(taskID int, fieldValuePairs []string) (bool, error) {
	t, err := findTaskById(taskID)
	if err != nil {
		return false, err
	}
	t.updateFieldValuePairs(fieldValuePairs)
	return true, nil
}

// AddTimerForTask adds a timer for a given task id and duration
func AddTimerForTask(taskID int, d time.Duration) (bool, error) {
	t, err := findTaskById(taskID)
	if err != nil {
		panic(err)
	}
	if hasRunningTimer(taskID) {
		return false, errDuplicateTimerForTask
	}

	startedAt := time.Now()
	expiresAt := startedAt.Add(d)

	DB.Save(&taskTimer{
		TaskID:    taskID,
		Message:   "Timer for '" + t.Name + "' finished!",
		StartedAt: startedAt,
		ExpiresAt: expiresAt,
	})
	return true, nil
}

func CancelTimerForTask(taskID int) {
	var tt taskTimer
	DB.One("TaskID", taskID, &tt)
	err := DB.DeleteStruct(&tt)
	if err != nil {
		panic(err)
	}
}

func findTaskById(id int) (*task, error) {
	var t task
	err := DB.One("ID", id, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func hasRunningTimer(taskID int) bool {
	var taskTimers []taskTimer
	_ = DB.Select(q.And(q.Eq("TaskID", taskID), q.Eq("Fired", false))).Find(&taskTimers)
	return len(taskTimers) != 0
}

func (tt *taskTimer) pushNotification() {
	gosxnotifier.NewNotification(tt.Message).Push()
}

func createRows() [][]string {
	var rows = [][]string{}
	var tasks []task
	err := DB.All(&tasks)
	if err != nil {
		fmt.Println("You still have no tasks added.")
		fmt.Println("Add the first one by typing 'add <task-name>'")
	}
	for _, tt := range tasks {
		row := []string{
			strconv.Itoa(tt.ID),
			tt.Name,
			coloredStatus(tt.Status),
			coloredPriority(tt.Priority),
			strings.Join(tt.Tags, ","),
			totalRunningTime(tt.ID).String(),
		}
		rows = append(rows, row)
	}
	return rows
}

func coloredPriority(p int) string {
	if p > 0 && p <= 2 {
		return color.GreenString(strconv.Itoa(p))
	} else if p < 5 {
		return color.YellowString(strconv.Itoa(p))
	}
	return color.RedString(strconv.Itoa(p))
}

func totalRunningTime(taskID int) time.Duration {
	total := time.Duration(0)
	var taskTimers []taskTimer
	err := DB.Find("TaskID", taskID, &taskTimers)
	if err != nil {
		return total
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
