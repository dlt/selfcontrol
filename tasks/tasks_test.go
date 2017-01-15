package tasks

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var testdbfile = filepath.Join(os.Getenv("HOME"), ".selfcontrol_test.db")

func TestMain(m *testing.M) {
	Init(testdbfile)
	m.Run()
	resetDB()
}

func resetDB() {
	err := os.Remove(testdbfile)
	if err != nil {
		panic(err)
	}
}

func TestAdd(t *testing.T) {
	Add("foo", make([]string, 0))
	var tsk task
	err := DB.One("Name", "foo", &tsk)
	if err != nil {
		t.Error(err)
	}
	if tsk.Name != "foo" {
		t.Error("expected task name", "foo", tsk.Name)
	}
	fieldValuePairs := []string{"pri:1", "status:done"}
	Add("bar", fieldValuePairs)
	err = DB.One("Name", "bar", &tsk)
	if err != nil {
		t.Error(err)
	}
	if tsk.Priority != 1 {
		t.Error("expected task priority", 1, tsk.Priority)
	}
	if tsk.Status != "DONE" {
		t.Error("expected task status", "done", tsk.Status)
	}
}

func TestDelete(t *testing.T) {
	Add("foo", make([]string, 0))
	var tsk task
	err := DB.One("Name", "foo", &tsk)
	if err != nil {
		t.Error(err)
	}
	if tsk.Name != "foo" {
		t.Error("expected task name", "foo", tsk.Name)
	}

	Delete(tsk.ID)
	err = DB.One("Name", "foo", &tsk)
	assert.Error(t, err, "not found error was expected")
}

func TestUpdate(t *testing.T) {
	Add("foo", make([]string, 0))
	var tsk task
	err := DB.One("Name", "foo", &tsk)
	if err != nil {
		t.Error(err)
	}
	if tsk.Name != "foo" {
		t.Error("expected task name", "foo", tsk.Name)
	}

	fieldValuePairs := []string{"pri:10", "status:done", "n:eggs", "t:foo"}
	UpdateFields(tsk.ID, fieldValuePairs)
	err = DB.One("ID", tsk.ID, &tsk)

	if tsk.Name != "eggs" {
		t.Error("expected task Name", "eggs", tsk.Name)
	}
	if tsk.Status != "DONE" {
		t.Error("expected task Status", "DONE", tsk.Status)
	}
	if tsk.Priority != 10 {
		t.Error("expected task Priority", 10, tsk.Priority)
	}
	tags := []string{"foo"}
	if !reflect.DeepEqual(tsk.Tags, tags) {
		t.Error("expected task Tags", tags, tsk.Tags)
	}

	fieldValuePairs = []string{"t:-foo"}
	UpdateFields(tsk.ID, fieldValuePairs)
	err = DB.One("ID", tsk.ID, &tsk)
	tags = make([]string, 0)
	if !reflect.DeepEqual(tsk.Tags, tags) {
		t.Error("expected task Tags", tags, tsk.Tags)
	}

}

func TestAddTags(t *testing.T) {
	name := "withtags"
	fieldValuePairs := []string{"t:foo,bar,eggs,biz,-biz"}

	Add(name, fieldValuePairs)
	var tsk task
	_ = DB.One("Name", name, &tsk)
	tags := []string{"foo", "bar", "eggs"}
	if !reflect.DeepEqual(tsk.Tags, tags) {
		t.Error("expected task Tags", tags, tsk.Tags)
	}
}

func TestAddTimer(t *testing.T) {
	name := "withtimer"
	fieldValuePairs := make([]string, 0)
	Add(name, fieldValuePairs)
	var tsk task
	_ = DB.One("Name", name, &tsk)

	duration, _ := time.ParseDuration("1m")

	AddTimerForTask(tsk.ID, duration)

	var tt taskTimer
	_ = DB.One("TaskID", tsk.ID, &tt)

	if tt.TaskID != tsk.ID {
		t.Error("expected task ID", tt.TaskID, tsk.ID)
	}
}

func TestCancelTimer(t *testing.T) {
	name := "withcanceledtimer"
	fieldValuePairs := make([]string, 0)
	Add(name, fieldValuePairs)
	var tsk task
	_ = DB.One("Name", name, &tsk)

	duration, _ := time.ParseDuration("1m")

	AddTimerForTask(tsk.ID, duration)

	var tt taskTimer
	_ = DB.One("TaskID", tsk.ID, &tt)

	if tt.TaskID != tsk.ID {
		t.Error("expected task ID", tt.TaskID, tsk.ID)
	}

	CancelTimerForTask(tsk.ID)
	err := DB.One("TaskID", tsk.ID, &tt)
	if err == nil {
		t.Error("expected task to have canceled timer", tt.TaskID, 0)
	}
}
