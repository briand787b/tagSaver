package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	db        *sqlx.DB
	dbWorkers = make(chan struct{}, 500)
)

type dbCredentials struct {
	Name     string `json:"dbName"`
	UserName string `json:"dbUserName"`
	Password string `json:"dbPassword"`
}

func init() {
	creds, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("could not read config file: %s", err)
	}

	var dbc dbCredentials
	if err := json.Unmarshal(creds, &dbc); err != nil {
		log.Fatalf("could not unmarshal credentials: %s", err)
	}

	dsn := fmt.Sprintf("%s:%s@/%s", dbc.UserName, dbc.Password, dbc.Name)
	db = sqlx.MustConnect("mysql", dsn)
}

// this should be run as a goroutine.  Therefore returning errors does
// not make sense
func saveTags(tags ...string) {
	// get token from chan to proceed
	dbWorkers <- struct{}{}

	q := `
		INSERT INTO tag
		(
			tag_number
		)
		VALUES
		`
	vals := make([]interface{}, len(tags))
	for i, tag := range tags {
		q += `(?),`
		vals[i] = tag
	}

	q = q[:len(q)-1]
	// fmt.Println("query: ", q)
	// tx, err := db.BeginTxx(ctx, nil)
	// if err != nil {
	// 	fmt.Println("could not save tags: %v", tags)
	// 	return
	// }

	r, err := db.ExecContext(ctx, q, vals...)
	if err != nil {
		fmt.Printf("could not save tags(%v): %s\n", tags, err)
		return
	}

	rows, err := r.RowsAffected()
	if err != nil {
		fmt.Printf("failed to save set(%v): could not get number of rows affected: %s\n", tags, err)
	}

	if rows < 1 {
		fmt.Println("WARNING: no rows affected for tag set: ", tags)
	}

	// dequeue worker to allow new workers to get tokens
	<-dbWorkers

	return
}
