package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const url = "http://localhost:8080/quotation"
const aspas = `"`

func main() {
	f, err := os.Create("quotation.txt")
	if err != nil {
		log.Println("Error creating file")
		os.Exit(0)
	}
	defer f.Close()
	f.WriteString("Dólar: " + strings.ReplaceAll(getDollarValue(), aspas, ""))
}

func getDollarValue() string {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	valueDollar, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	return string(valueDollar)
}
