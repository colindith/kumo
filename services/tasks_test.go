package services

import (
	"testing"
)

func test_crawler(t *testing.T){
	err := GetData("2330")
	if err != nil {
		t.Error("got error", err)
	}
}