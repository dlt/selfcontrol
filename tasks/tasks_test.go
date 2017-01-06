package tasks

import (
        "testing"
        "path/filepath"
        "os"
)

var (
        testdbfile string = filepath.Join(os.Getenv("HOME"), ".selfcontrol_test.db")
)

func TestMain(m *testing.M) {
        Init(testdbfile)
        m.Run()
        resetDB()
}

func resetDB()  {
        err := os.Remove(testdbfile)
        if  err != nil {
                panic(err)
        }
}

func TestAdd(t *testing.T)  {
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
