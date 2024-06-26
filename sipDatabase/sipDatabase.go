package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "sip_box"
)

// Globals here
var username, password, sip_password string
var extension, id int

func main() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// insert
	// hardcoded
	var newNumber int
	query := `SELECT * FROM users;`
	rows, e := db.Query(query)
	CheckError(e)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &extension, &sip_password)
		CheckError(err)

		fmt.Println(id, username, password, extension, sip_password)
		newNumber = extension + 1
	}
	insertQry := `INSERT INTO users (username, password, extension, sip_password) VALUES ('bruggy','bruggerton',$1,'bruggerton') RETURNING username, password, extension, sip_password;`
	insrtStmt, e := db.Exec(insertQry, newNumber)
	CheckError(e)
	defer insrtStmt.Close()
	for insrtStmt.Next() {
		err = insrtStmt.Scan(&id, &username, &password, &extension, &sip_password)
		CheckError(err)

		fmt.Println(id, username, password, extension, sip_password)
	}

	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
