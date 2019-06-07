package services

import (
	"fmt"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func stockPrice(server *machinery.Server) func() {
	return func() {
		fmt.Println("stockPrice task ~~~~~~~~~~~")
		longRunningTask := &tasks.Signature{
			Name: "long_running_task",
		}
		asyncResult, err := server.SendTask(longRunningTask)
		if err != nil {
			// failed to send the task
			// do something with the error
		}
		fmt.Println("asyncResult: ", asyncResult)
	}
}

