package main

import (
	"fmt"
	"net/http"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func main() {
	// init worker
	var cnf = &config.Config{
		Broker:        "amqp://taco:p@ss1234@taco-rabbitmq:5672/taco_vhost",
		DefaultQueue:  "machinery_tasks",
		ResultBackend: "amqp://taco:p@ss1234@taco-rabbitmq:5672/taco_vhost",
		AMQP: &config.AMQPConfig{
			Exchange:     "machinery_exchange",
			ExchangeType: "direct",
			BindingKey:   "machinery_task",
		},
	}

	server, err := machinery.NewServer(cnf)
	if err != nil {
		// do something with the error
	}

	// periodic work
	c := cron.New()
	c.AddFunc("5 * * * * *", periodicFunc(server))
	// c.AddFunc("5 * * * * *", stockCrawler(server))
	c.Start()

	r := gin.Default()

	r.GET("/", index(server))

	r.Run(":8001")
}

func index(server *machinery.Server) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.String(http.StatusOK, "hello, world")
		signature := &tasks.Signature{
			Name: "add",
			Args: []tasks.Arg{
				{
					Type:  "int64",
					Value: 1,
				},
				{
					Type:  "int64",
					Value: 1,
				},
			},
		}

		asyncResult, err := server.SendTask(signature)
		if err != nil {
			// failed to send the task
			// do something with the error
		}
		fmt.Println("asyncResult: ", asyncResult)
	}
}
func periodicFunc(server *machinery.Server) func() {
	return func() {
		fmt.Println("Every minute on the 25 sec")
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
