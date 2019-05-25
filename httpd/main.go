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
	c.AddFunc("*/5 * * * * *", periodicFunc)
	c.Start()

	// 初始化引擎
	r := gin.Default()
	// 注册一个路由和处理函数
	r.GET("/", index(server))
	// 绑定端口，然后启动应用
	r.Run(":8001")
}

/**
* 根请求处理函数
* 所有本次请求相关的方法都在 context 中，完美
* 输出响应 hello, world
 */

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

func periodicFunc() {
	fmt.Println("Every minute on the 25 sec")
}
