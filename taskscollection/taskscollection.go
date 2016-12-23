package taskscollection

import (
	"encoding/json"
	"fmt"
	"os"
	"errors"
	"strconv"
	"github.com/olekukonko/tablewriter"
	"github.com/HouzuoGuo/tiedot/db"
	_ "github.com/HouzuoGuo/tiedot/dberr"
	_ "io/ioutil"
	_ "log"
        _ "reflect"
	_ "time"
	_ "strings"
	_ "github.com/deckarep/gosx-notifier"
	_ "github.com/deckarep/gosx-notifier"
)

var (
        myDBDir 		  = "/Users/dlt/.selfcontrol"
	collectionName		  = "Tasks"
        ErrNoSuchTask             = errors.New("no such task")
	tasksCollection		  *db.Col
)

type Task map[string]interface{}

func init()  {
	tasksDB, err := db.OpenDB(myDBDir)
	if err != nil {
		panic(err)
	}
	tasksDB.Create(collectionName)
	tasksCollection = tasksDB.Use(collectionName)
}

func Create(name string) {
	task := map[string]interface{}{
		"name":               name,
		"status":             "TODO",
	}
	docID, err := tasksCollection.Insert(task)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created doc with id: %s", docID)
}
func Print()  {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "name", "status"})
	for _, row := range createRows() {
		table.Append(row)
	}
	table.Render()
}

func createRows() [][]string {
	var rows = [][]string{}
	tasksCollection.ForEachDoc(func (id int, raw []byte) (willMoveOn bool) {
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

/*func ellapsedTime(t Task) string {
	timer := timers[t.ID]
	seconds := time.Now().Sub(timer.StartedAt).Seconds()
	if t.Status != "DOING" {
		return "-"
	}
	return strconv.FormatFloat(seconds, 'E', -1, 32)
}

type taskTimer struct {
	Timer     *time.Timer
	StartedAt time.Time
	Task      *Task
}

timers                    = make(map[int]taskTimer)
func (tc *TasksCollection) UpdateStatus(taskID int, status string) (bool, error) {
	task, err := tc.Find(taskID)
        if err != nil {
                return false, ErrNoSuchTask
        }
	task.Status = strings.ToUpper(status)
	tc.persist()
	return true, nil
}

func (tc *TasksCollection) Remove(id int) error {
        var newTasks []Task
        var noSuchTask = true
	for i := 0; i < len(tc.tasks); i++ {
		t := tc.tasks[i]
		if t.ID != id {
			newTasks = append(newTasks, t)
		} else {
                        noSuchTask = false
                }
	}
        if noSuchTask {
                return ErrNoSuchTask
        }
	tc.tasks = newTasks
        tc.Outdated = true
        return nil
}

func (tc *TasksCollection) StartTimerForTask(taskID int, d time.Duration) {
	task, err := tc.Find(taskID)
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
}*/
