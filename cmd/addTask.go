// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
	"log"
	"github.com/spf13/cobra"
)

type Task struct {
	Id 		int
	Name	string
	Status	string
	TotalTimeInSeconds int
}

var taskName string
var taskStatus string

var tasks []Task
var lastId int

// addTaskCmd represents the addTask command
var addTaskCmd = &cobra.Command{
	Use:   "addTask",
	Short: "Creates a task",
	Long: `Creates a task.`,

	Run: func(cmd *cobra.Command, args []string) {

		task := Task{
			Id: 1,
			Name: taskName,
			Status: taskStatus,
			TotalTimeInSeconds: 0,
		}

		save(task)
	},
}

func save(task Task)  {
		if isDatabaseFileEmpty() {
			tasks = []Task{ task }
		} else {
			readDatabaseFile()
		}

		tasks = append(tasks, task)
		jsonString := toJSON(tasks)

		writeToDatabaseFile(jsonString)
}

func readDatabaseFile() {
	raw, err := ioutil.ReadFile(TasksFile)
	if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
	}

	json.Unmarshal(raw, &tasks)
}

func writeToDatabaseFile(jsonString []byte)  {
	f, err := os.Create(TasksFile)
	defer f.Close()

	if err != nil {
		fmt.Println("error:", err)
	}

	bytesWritten, err := f.Write(jsonString)
	if err != nil {
		fmt.Println("error:", err)
	}
	f.Sync()

	fmt.Println("writing to file: ", TasksFile)
	fmt.Println("bytesWritten: ", bytesWritten)
}

func toJSON(tasks []Task) []byte {
		j, err := json.Marshal(tasks)
		if err != nil {
			fmt.Println("error:", err)
		}

		fmt.Println("Writing tasks to file: ", j)
		return j
}

func isDatabaseFileEmpty() bool {
	file, err := os.Open(TasksFile)

	if err != nil {
	    log.Fatal(err)
	}
	f, err := file.Stat()
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println("filesize: ", f.Size())

	return f.Size() == 0
}

func init() {
	RootCmd.AddCommand(addTaskCmd)
	addTaskCmd.Flags().StringVarP(&taskName, "name", "n", "no-name", "Task name")
	addTaskCmd.Flags().StringVarP(&taskStatus, "status", "s", "todo", "Task status")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addTaskCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addTaskCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
