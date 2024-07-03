package databaser

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host       = "localhost"
	port       = 5432
	user       = "postgres"
	dbPassword = "postgres"
	dbname     = "sip_box"
)

// Globals here
var sip_password string
var extensions int

func PullRegistry() map[int]string {
	registry := make(map[int]string)
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode-disable", host, port, user, dbPassword, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// select registration creds
	query := `SELECT extension, sip_password FROM USERS;`
	rows, e := db.Query(query)
	CheckError(e)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&extension, &sip_password)
		CheckError(err)

		registry[extension] = sip_password
	}
	fmt.Println("Pulled registry") // maybe could add a line about when it was pulled and by who?
	return registry
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
