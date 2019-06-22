package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kumo/httpd/handler/stock"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/streadway/amqp"
	// "github.com/RichardKnop/machinery/v1"
)

var db *gorm.DB
var err error

// Stock Price from Yahoo
func StockCrawler() error {
	// Move this init DB code to some where out....
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
		return err
	}
	defer db.Close()

	// TODO: Add a table for all stock code
	// TODO: run through all Taiwan stock here
	fmt.Println("Starting crawl stock........")
	codes := []string{"2330"}
	count := 0
	for _, code := range codes {
		err := GetData(code, db)
		if err != nil {
			fmt.Println(err.Error())
		}
		count++
	}
	return nil
}

func GetData(code string, db *gorm.DB) error {
	// BroadcastMessage("kumo_broadcast", "yayayay!!message from kumo~~~~~")

	// get response
	resp, err := http.Get("https://tw.quote.finance.yahoo.net/quote/q?type=tick&perd=1m&mkt=10&sym=" + code)
	if err != nil {
		fmt.Println("http error: ", err.Error())
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// truncate respnse & leave the content in bracket
	var buffer bytes.Buffer
	in_json := false
	for _, b := range body {
		if b == 41 { // ")" = 41
			in_json = false
		}
		if in_json == true {
			// fmt.Print(b)
			buffer.WriteByte(b)
		}
		if b == 40 { // "(" = 40
			in_json = true
		}
	}

	// parse the []byte response
	var stockData interface{}
	err2 := json.Unmarshal(buffer.Bytes(), &stockData)
	if err2 != nil {
		fmt.Println(err2)
		return err2
	}
	data := stockData.(map[string]interface{})
	// fmt.Print(data)
	mkt := data["mkt"].(string)
	id := data["id"].(string)
	mem := data["mem"].(map[string]interface{})
	iTickArray := data["tick"].([]interface{})

	// Get or Create the Product from db
	var product stock.Product
	err = db.Where("Code = ?", id).First(&product).Error
	if gorm.IsRecordNotFoundError(err) {
		product = stock.Product{Code: id}
		db.Create(&product)
	}
	// fmt.Println("product: ", product)

	totalSaveCount := 0
	for _, iTick := range iTickArray {
		tick, _ := iTick.(map[string]interface{})
		intTime := int(tick["t"].(float64))
		tickTime := intToTime(intTime)
		price := tick["p"].(float64)
		volume := tick["v"].(float64)

		// Check if result already in db
		var priceObj stock.Price
		err := db.Where("time = ?", tickTime).Where("product_id = ?", product.ID).First(&priceObj).Error
		if err == nil {
			// duplicated result, skip
			// continue
		} else if !gorm.IsRecordNotFoundError(err) {
			fmt.Println("get price error: ", err.Error())
			// return err
		}

		// Prepare the price object
		priceObj = stock.Price{
			Product: product,
			Time:    tickTime,
			Price:   price,
			Volume:  uint(volume),
		}
		//save the Price object
		if err := db.Set("gorm:save_associations", false).Create(&priceObj).Error; err != nil {
			fmt.Println("create price error: ", err.Error())
			// return err
		}

		totalSaveCount++
		// TODO: push the new price info to broker
		msg := make(map[string]interface{})
		msg["product_id"] = priceObj.Product.ID
		msg["time"] = priceObj.Time
		msg["price"] = priceObj.Price
		msg["volume"] = priceObj.Volume

		b, err := json.Marshal(msg)
		if err != nil {
			fmt.Print("parse price object into json error", err.Error())
		}
		BroadcastMessage("kumo_broadcast", b)

		// var p stock.Price
		// json.Unmarshal(b, &p)
		fmt.Println(b)
		// fmt.Println(p.Price)
	}

	fmt.Println(mkt, id, mem)
	fmt.Println("totalSaveCount: ", totalSaveCount)
	return nil
}

func intToTime(i int) time.Time {
	minutes := i % 100
	i /= 100
	hours := i % 100
	i /= 100
	date := i % 100
	i /= 100
	month := i % 100
	i /= 100
	year := i
	return time.Date(year, time.Month(month), date, hours, minutes, 0, 0, time.Local)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func BroadcastMessage(exchange string, msg []byte) {
	rabbitmqHost := os.Getenv("RABBITMQ_HOST")
	rabbitmqUser := os.Getenv("RABBITMQ_DEFAULT_USER")
	rabbitmqPass := os.Getenv("RABBITMQ_DEFAULT_PASS")
	rabbitmqVhost := os.Getenv("RABBITMQ_DEFAULT_VHOST")
	rabbitmqDial := fmt.Sprintf("amqp://%s:%s@%s:5672/%s", rabbitmqUser, rabbitmqPass, rabbitmqHost, rabbitmqVhost)
	conn, err := amqp.Dial(rabbitmqDial)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// body := bodyFrom(os.Args)
	// body := "broadcast msg ~~~ ~~~"
	err = ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", msg)
}
