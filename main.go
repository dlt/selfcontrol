package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/dlt/selfcontrol/tasks"
	"github.com/fatih/color"
)

var errInvalidArgumentList = errors.New("invalid argument list")
var errInvalidNumericArgument = errors.New("invalid numeric argument")
var shell = ishell.New()


func main() {
	banner := color.YellowString("The greatest conquest is ") + color.RedString("self-control.")
	shell.Println(banner)
	shell.Register("list", listTasks)
	shell.Register("add", addTask)
	shell.Register("delete", deleteTask)
	shell.Register("update", updateTask)
	shell.Register("edit", updateTask)
	shell.Register("timer", addTimerForTask)
	shell.Register("cancel", cancelTimerForTask)
	shell.Register("exit", exit)
	shell.Register("quit", exit)
	shell.SetHomeHistoryPath(".ishell_history")
	tasks.Init(getDBFile())
	tasks.Print()
	shell.Start()
}

func getDBFile() string {
	args := os.Args
	if len(args) == 1 {
		homedir := os.Getenv("HOME")
		return filepath.Join(homedir, ".selfcontrol.db")
	}
	return args[1]
}

func listTasks(args ...string) (string, error) {
	tasks.Print()
	return "", nil
}

func exit(args ...string) (string, error) {
	shell.Stop()
	return "", nil
}

func addTask(args ...string) (string, error) {
	if len(args) < 1 {
		return "", errInvalidArgumentList
	}
	name := args[0]
	if len(args) == 1 {
		args = append(args, "pri:0")
	}
	err := tasks.Add(name, args[1:])
	if err != nil {
		return "", err
	}
	tasks.Print()
	return "task created", nil
}

func deleteTask(args ...string) (string, error) {
	if len(args) < 1 {
		return "", errInvalidArgumentList
	}
	ids := make([]int, 0)
	if strings.Index(args[0], ",") > 0 {
		strs := strings.Split(args[0], ",")
		for _, str := range strs {
			id, err := strconv.Atoi(str)
			if err != nil {
				return args[0], errInvalidNumericArgument
			}
			ids = append(ids, id)
		}
	} else {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return args[0], errInvalidNumericArgument
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		tasks.Delete(id)
	}
	tasks.Print()
	return "", nil
}

func updateTask(args ...string) (string, error) {
	if len(args) < 2 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}
	_, err = tasks.UpdateFields(id, args[1:])
	if err != nil {
		return "couldn't update task", err
	}
	tasks.Print()
	return "", nil
}

func addTimerForTask(args ...string) (string, error) {
	if len(args) != 2 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}
	timeInMinutes, err := time.ParseDuration(args[1] + "m")
	if err != nil {
		return args[1], errInvalidNumericArgument
	}
	_, err = tasks.AddTimerForTask(id, timeInMinutes)
	if err != nil {
		return "", err
	}
	return "", nil
}

func cancelTimerForTask(args ...string) (string, error) {
	if len(args) != 1 {
		return "", errInvalidArgumentList
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return args[0], errInvalidNumericArgument
	}
	tasks.CancelTimerForTask(id)
	tasks.Print()
	return "", nil
}
