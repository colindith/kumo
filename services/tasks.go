package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Stock Price from Yahoo
func StockCrawler(codes ...string) (int, error) {
	count := 0
	for _, code := range codes {
		err := GetData(code)
		count++
	}
	return count, nil
}

func GetData(code string) {
	// get response
	resp, err := http.Get("https://tw.quote.finance.yahoo.net/quote/q?type=tick&perd=1m&mkt=10&sym=" + code)
	if err != nil {
		fmt.Println("http error: ", err.Error())
		return
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
	var stock interface{}
	err2 := json.Unmarshal(buffer.Bytes(), &stock)
	if err2 != nil {
		fmt.Println(err2)
	}
	data := stock.(map[string]interface{})
	// fmt.Print(data)
	mkt := data["mkt"].(string)
	id := data["id"].(string)
	mem := data["mem"].(map[string]interface{})
	iTickArray := data["tick"].([]interface{})
	for i, iTick := range iTickArray {
		tick, _ := iTick.(map[string]interface{})
		intTime := int(tick["t"].(float64))
		tickTime := intToTime(intTime)
		price := tick["p"].(float64)
		volume := tick["v"].(float64)
		fmt.Println(i, id, tickTime, price, volume)
	}

	fmt.Println(mkt, id, mem)
	return nil
}

func SaveData() {

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
