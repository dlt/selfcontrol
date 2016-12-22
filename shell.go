package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type Task struct {
	Id                 int
	Name               string
	Status             string
	TotalTimeInSeconds int
}

var (
	InvalidArgumentList                        = errors.New("invalid argument list")
	InvalidNumericArgument                     = errors.New("invalid numeric argument")
	tasksFilepath          string              = "/Users/dlt/golang/src/SelfControl/tasks.json"
	timers                 map[int]*time.Timer = make(map[int]*time.Timer)
	tasks                  []Task
	lastId                 int
)

func main() {
	shell := ishell.New()
	shell.Println("The greatest conquest is self–control")

	shell.Register("list", listTasks)
	shell.Register("add", addTask)
	shell.Register("delete", deleteTask)
	shell.Register("start", start)

	shell.Start()
}

// Print all tasks in a ASCII table
func listTasks(args ...string) (string, error) {
	readDatabaseFile()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "name", "status", "time"})

	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()

	return "", nil
}

// Creates a new task given name and status
func addTask(args ...string) (string, error) {
	if len(args) < 2 {
		return "", InvalidArgumentList
	}
	name := args[0]
	status := args[1]
	saveTask(name, status)
	return "task created", nil
}

// Deletes a task with given id
func deleteTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", InvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], InvalidNumericArgument
	}
	if tasks == nil {
		readDatabaseFile()
	}
	tasks := removeTaskWithId(tasks, id)
	persist(tasks)
	listTasks("")
	return "", nil
}

// Starts a timer for a given task id
func start(args ...string) (string, error) {
	if len(args) != 2 {
		return "", InvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], InvalidNumericArgument
	}

	timeInMinutes, err := time.ParseDuration(args[1] + "s")
	if err != nil {
		return args[1], InvalidNumericArgument
	}

	task := findTaskWithId(id)
	startTimerForTask(task, timeInMinutes)
	return "", nil
}

func startTimerForTask(t *Task, d time.Duration) {
	timer := time.NewTimer(d)
	timers[t.Id] = timer

	go func() {
		<-timer.C
		fmt.Println("Timer expired")
	}()
}

func removeTaskWithId(tasks []Task, id int) []Task {
	var newTasks []Task
	for _, t := range tasks {
		if t.Id != id {
			newTasks = append(newTasks, t)
		}
	}
	return newTasks
}

func findTaskWithId(id int) *Task {
	for _, t := range getTasks() {
		if t.Id == id {
			return &t

		}
	}
	return nil
}

func getTasks() []Task {
	if isDatabaseFileEmpty() {
		return make([]Task, 0)
	} else {
		readDatabaseFile()
		return tasks
	}
}

func createRows() [][]string {
	var rows = [][]string{}

	for _, t := range getTasks() {
		row := []string{strconv.Itoa(t.Id), t.Name, t.Status, strconv.Itoa(t.TotalTimeInSeconds)}
		rows = append(rows, row)
	}

	return rows
}

func readDatabaseFile() {
	raw, err := ioutil.ReadFile(tasksFilepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	json.Unmarshal(raw, &tasks)
}

func saveTask(name string, status string) {
	task := Task{
		Id:                 0,
		Name:               name,
		Status:             status,
		TotalTimeInSeconds: 0,
	}

	tasks = getTasks()
	task.Id = maxId(tasks) + 1

	tasks = append(tasks, task)
	persist(tasks)
}

func persist(tasks []Task) {
	writeToDatabaseFile(toJSON(tasks))
}

func maxId(tasks []Task) int {
	id := 0

	for _, t := range tasks {
		if t.Id > id {
			id = t.Id
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
	f.Sync()
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
