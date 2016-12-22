package main

import (
	"github.com/abiosoft/ishell"
	"strconv"
	"time"
	"errors"
	"github.com/dlt/selfcontrol/taskscollection"
)

var (
        ErrInvalidArgumentList    = errors.New("invalid argument list")
        ErrInvalidNumericArgument = errors.New("invalid numeric argument")
        ErrNoSuchTask             = errors.New("no such task")
	tasks                     = taskscollection.TasksCollection{Outdated:false}
	lastID                    int
)

func main() {
	shell := ishell.New()
	shell.Println("The greatest conquest is self–control")

        listTasks("")
	shell.Register("list", listTasks)
	shell.Register("add", addTask)
	shell.Register("delete", deleteTask)
	shell.Register("start", startTask)
	shell.Register("status", setTaskStatus)

	shell.Start()
}

// Print all tasks in a ASCII table
func listTasks(args ...string) (string, error) {
        tasks.Load()
	tasks.Print()
	return "", nil
}

func setTaskStatus(args ...string) (string, error) {
	if len(args) != 2 {
		return "", ErrInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], ErrInvalidNumericArgument
	}
	status := args[1]
	_, err = tasks.UpdateStatus(id, status)
	if err != nil {
		return "couldn't update task", err
	}
	return "", nil
}

// Creates a new task given name and status
func addTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", ErrInvalidArgumentList
	}
	name := args[0]
	tasks.Create(name)
	tasks.Print()
	return "task created", nil
}

// Deletes a task with given id
func deleteTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", ErrInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], ErrInvalidNumericArgument
	}
        if err = tasks.Remove(id); err != nil {
                return "", err
        }
	listTasks("")
	return "", nil
}

// Starts a timer for a given task id
func startTask(args ...string) (string, error) {
	if len(args) != 2 {
		return "", ErrInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], ErrInvalidNumericArgument
	}

	timeInMinutes, err := time.ParseDuration(args[1] + "s")
	if err != nil {
		return args[1], ErrInvalidNumericArgument
	}
	tasks.StartTimerForTask(id, timeInMinutes)
	return "", nil
}
