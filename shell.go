package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/deckarep/gosx-notifier"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// Task struct
type Task struct {
	ID                 int
	Name               string
	Status             string
	TotalTimeInSeconds int
}

type taskTimer struct {
	Timer     *time.Timer
	StartedAt time.Time
	Task      *Task
}

var (
	errInvalidArgumentList    = errors.New("invalid argument list")
	errInvalidNumericArgument = errors.New("invalid numeric argument")
	tasksFilepath             = "/Users/dlt/golang/src/SelfControl/tasks.json"
	timers                    = make(map[int]taskTimer)
	tasks                     []Task
	lastID                    int
)

func main() {
	shell := ishell.New()
	shell.Println("The greatest conquest is self–control")

	shell.Register("list", listTasks)
	shell.Register("add", addTask)
	shell.Register("delete", deleteTask)
	shell.Register("start", startTask)
	shell.Register("status", setTaskStatus)

	shell.Start()
}

// Print all tasks in a ASCII table
func listTasks(args ...string) (string, error) {
	readDatabaseFile()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "name", "status", "time", "ellapsed time"})

	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()

	return "", nil
}

func setTaskStatus(args ...string) (string, error) {
	if len(args) != 2 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}
	status := args[1]
	task := findTaskWithID(id)
	task.setStatus(status)
	return "", nil
}

// Creates a new task given name and status
func addTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", errInvalidArgumentList
	}
	name := args[0]
	saveTask(name)
	return "task created", nil
}

// Deletes a task with given id
func deleteTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}
	if tasks == nil {
		readDatabaseFile()
	}
	removeTaskWithID(id)
	persist()
	listTasks("")
	return "", nil
}

// Starts a timer for a given task id
func startTask(args ...string) (string, error) {
	if len(args) != 2 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}

	timeInMinutes, err := time.ParseDuration(args[1] + "s")
	if err != nil {
		return args[1], errInvalidNumericArgument
	}
	task := findTaskWithID(id)
	task.setStatus("DOING")
	task.startTimerForTask(timeInMinutes)
	return "", nil
}

func (t *Task) setStatus(status string) {
	t.Status = status
	removeTaskWithID(t.ID)
	tasks = append(tasks, *t)
	persist()
}

func (t *Task) startTimerForTask(d time.Duration) {
	timer := time.NewTimer(d)
	timers[t.ID] = taskTimer{
		Task:      t,
		Timer:     timer,
		StartedAt: time.Now(),
	}
	message := "Timer for '" + t.Name + "' finished!"
	go func() {
		<-timer.C
		pushNotification(message)
	}()
}

func removeTaskWithID(id int) {
	var newTasks []Task
	for _, t := range tasks {
		if t.ID != id {
			newTasks = append(newTasks, t)
		}
	}
	tasks = newTasks
}

func findTaskWithID(id int) *Task {
	for _, t := range getTasks() {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

func getTasks() []Task {
	if isDatabaseFileEmpty() {
		return make([]Task, 0)
	}
	readDatabaseFile()
	return tasks
}

func createRows() [][]string {
	var rows = [][]string{}

	for _, t := range getTasks() {
		row := []string{
			strconv.Itoa(t.ID),
			t.Name,
			t.Status,
			strconv.Itoa(t.TotalTimeInSeconds),
			ellapsedTime(t),
		}
		rows = append(rows, row)
	}

	return rows
}

func ellapsedTime(t Task) string {
	timer := timers[t.ID]
	minutes := timer.StartedAt.Sub(time.Now()).Minutes()
	if t.Status != "DOING" {
		return "-"
	}
	return strconv.FormatFloat(minutes, 'E', -1, 32)
}

func readDatabaseFile() {
	raw, err := ioutil.ReadFile(tasksFilepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(raw, &tasks)
}

func saveTask(name string) {
	task := Task{
		ID:                 0,
		Name:               name,
		Status:             "TODO",
		TotalTimeInSeconds: 0,
	}

	tasks = getTasks()
	task.ID = maxID(tasks) + 1

	tasks = append(tasks, task)
	persist()
}

func persist() {
	writeToDatabaseFile(toJSON(tasks))
}

func maxID(tasks []Task) int {
	id := 0

	for _, t := range tasks {
		if t.ID > id {
			id = t.ID
		}
	}
	return id
}

func writeToDatabaseFile(jsonString []byte) {
	f, err := os.Create(tasksFilepath)
	defer f.Close()

	if err != nil {
		fmt.Println("error:", err)
	}

	_, err = f.Write(jsonString)
	if err != nil {
		fmt.Println("error:", err)
	}
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

func pushNotification(message string) {
	gosxnotifier.NewNotification(message).Push()
}
