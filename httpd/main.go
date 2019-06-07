package main

import (
	"fmt"
	"net/http"
	// "kumo/services"
	"os"
	"kumo/httpd/handler/stock"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB
var err error

func main() {
  // Init DB
	dbInfoStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("KUMO_POSTGRES_HOST"),
		os.Getenv("KUMO_POSTGRE_PORT"),
		os.Getenv("KUMO_POSTGRES_USERNAME"),
		os.Getenv("KUMO_POSTGRES_DB"),
		os.Getenv("KUMO_POSTGRES_PASSWORD"),
	)
	fmt.Println("dbInfoStr", dbInfoStr)
	db, err = gorm.Open("postgres", dbInfoStr)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	db.AutoMigrate(&stock.Product{})
	db.AutoMigrate(&stock.Price{})
	// db.AutoMigrate(&user.User{})

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
	// c.AddFunc("5 * * * * *", services.StockCrawler(server, db))
	c.AddFunc("35 * * * * *", periodicCrawlerFunc(server))
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
func periodicCrawlerFunc(server *machinery.Server) func() {
	return func() {
		fmt.Println("periodicCrawlerFunc")
		crawlerTask := &tasks.Signature{
			Name: "stock_crawler",
		}
		asyncResult, err := server.SendTask(crawlerTask)
		if err != nil {
			// failed to send the task
			// do something with the error
		}
		fmt.Println("asyncResult: ", asyncResult)
	}
}