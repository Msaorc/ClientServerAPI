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

const url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const nameDB = "./database/clientserver.db"
const sqlCreate = `CREATE TABLE dollarQuotation (
					Code TEXT,
					Codein TEXT,    
					Name TEXT,
					High TEXT,
					Low TEXT,
					VarBid TEXT, 
					PctChange TEXT,
					Bid TEXT,
					Ask TEXT, 
					Timestamp TEXT, 
					CreateDate TEXT);`
const sqlInsert = `INSERT INTO dollarQuotation (Code, Codein, Name, High, Low, VarBid, PctChange, Bid, Ask, TimeStamp, CreateDate)
                          VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

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
	dolarQuotation := findQuotation()
	db := createDatabase()
	saveQuotation(db, dolarQuotation)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dolarQuotation.Usdbrl.Bid)
}

func findQuotation() DollarQuotation {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
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
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	var dollarQuotation DollarQuotation
	json.Unmarshal(body, &dollarQuotation)
	return dollarQuotation
}

func saveQuotation(db *sql.DB, quotation DollarQuotation) {
	ctx := context.Background()
	stmt, err := db.Prepare(sqlInsert)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(quotation.Usdbrl.Code, quotation.Usdbrl.Codein, quotation.Usdbrl.Name, quotation.Usdbrl.High,
		quotation.Usdbrl.Low, quotation.Usdbrl.VarBid, quotation.Usdbrl.PctChange, quotation.Usdbrl.Bid,
		quotation.Usdbrl.Ask, quotation.Usdbrl.Timestamp, quotation.Usdbrl.CreateDate)
	if err != nil {
		panic(err)
	}
	select {
	case <-time.After(10 * time.Millisecond):
		log.Println("Quote successfully saved")
	case <-ctx.Done():
		log.Println("Timeout when saving quote")
	}
}

func createDatabase() *sql.DB {
	if !isExistsDB(nameDB) {
		_, err := os.Create(nameDB)
		if err != nil {
			panic(err)
		}
		db := openDB()
		err = createTable(db)
		if err != nil {
			panic(err)
		}
		return db
	}
	return openDB()
}

func createTable(db *sql.DB) error {
	stmt, err := db.Prepare(sqlCreate)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func isExistsDB(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", nameDB)
	if err != nil {
		panic(err)
	}
	return db
}
