package main

import (
	"errors"
	"github.com/abiosoft/ishell"
	"github.com/dlt/selfcontrol/tasks"
	"strconv"
	"time"
)

var ErrInvalidArgumentList = errors.New("invalid argument list")
var ErrInvalidNumericArgument = errors.New("invalid numeric argument")

func main() {
	shell := ishell.New()
	shell.Println("The greatest conquest is self–control")

	shell.Register("list", listTasks)
	shell.Register("add", addTask)
	shell.Register("delete", deleteTask)
	shell.Register("update", updateTask)
	shell.Register("timer", addTimerForTask)

	tasks.Print()
	shell.Start()
}

// Print all tasks in a ASCII table
func listTasks(args ...string) (string, error) {
	tasks.Print()
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
	tasks.Delete(id)
	tasks.Print()
	return "", nil
}

func updateTask(args ...string) (string, error) {
	if len(args) < 2 {
		return "", ErrInvalidArgumentList
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], ErrInvalidNumericArgument
	}

	_, err = tasks.UpdateFields(id, args[1:])
	if err != nil {
		return "couldn't update task", err
	}
	tasks.Print()
	return "", nil
}

// Adds a timer for a given task id
func addTimerForTask(args ...string) (string, error) {
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

	_, err = tasks.AddTimerForTask(id, timeInMinutes)
	if err != nil {
		return "", err
	}

	tasks.Print()
	return "", nil
}
