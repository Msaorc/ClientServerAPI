package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const urlRequest = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const nameDB = "./database/clientserver.db"
const sqlCreate = `CREATE TABLE dollarQuotation (
	                code TEXT, 
					codein TEXT,
					name TEXT,
					high TEXT,
					low TEXT,
					varbid TEXT,
					pctchange TEXT,
					bid TEXT,
					ask TEXT,
					timestamp TEXT,
					createdate TEXT );`
const sqlInser = `INSERT INTO dollarQuotation (code, codein, name, high, low, varbid, pctchange, bid, ask, timestamp, createdate) VALUES 
                   ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

type DollarQuotation struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/quotation", quotation)
	http.ListenAndServe(":8080", mux)
}

func quotation(w http.ResponseWriter, r *http.Request) {
	quotation := quotationRequest()
	saveToDatabase(&quotation)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotation.Usdbrl.Bid)
}

func quotationRequest() DollarQuotation {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", urlRequest, nil)
	if err != nil {
		panic(err)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	resJson, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	var dollarQuotation DollarQuotation
	err = json.Unmarshal(resJson, &dollarQuotation)
	if err != nil {
		panic(err)
	}
	return dollarQuotation
}

func saveToDatabase(quotation *DollarQuotation) {
	db, err := sql.Open("sqlite3", nameDB)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if !IsExistsDatabase() {
		os.Create(nameDB)
		err = createTable(db)
		if err != nil {
			panic(err)
		}
	}
	insertQuotation(db, quotation)
}

func IsExistsDatabase() bool {
	_, err := os.Stat(nameDB)
	return os.IsNotExist(err)
}

func createTable(db *sql.DB) error {
	stmt, err := db.Prepare(sqlCreate)
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func insertQuotation(db *sql.DB, quotation *DollarQuotation) error {
	ctx := context.Background()
	stmt, err := db.Prepare(sqlInser)
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec(quotation.Usdbrl.Code, quotation.Usdbrl.Codein, quotation.Usdbrl.Name, quotation.Usdbrl.High, quotation.Usdbrl.Low, quotation.Usdbrl.VarBid,
		quotation.Usdbrl.PctChange, quotation.Usdbrl.Bid, quotation.Usdbrl.Ask, quotation.Usdbrl.Timestamp, quotation.Usdbrl.CreateDate)
	if err != nil {
		return err
	}
	select {
	case <-time.After(3000 * time.Microsecond):
		log.Print("Quote successfully saved")
	case <-ctx.Done():
		log.Print("No action was taken, error: Timeout")
	}
	return nil
}
