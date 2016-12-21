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
	"os"
	"strconv"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		generateTable()
	},
}

func generateTable()  {
		readDatabaseFile()
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"id", "name", "status", "time"})

		for _, row := range createRows() {
			table.Append(row)
		}
		table.Render()
}

func createRows() [][]string {
	var rows = [][]string{}

	for _, t := range tasks {
		row := []string{strconv.Itoa(t.Id), t.Name, t.Status, strconv.Itoa(t.TotalTimeInSeconds)}
		rows = append(rows, row)
	}

	return rows
}

func init() {
	RootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
