package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("GO MYSQL START")
	db, err := sql.Open("mysql","root:pass@tcp(sql1:3306)/")

	if err != nil {
		panic(err)
	}

	defer db.Close()
	fmt.Println("Success open database")

	_,err = db.Exec("CREATE DATABASE IF NOT EXISTS CurrencyConv")
	if err != nil {
		panic(err)
	}

	fmt.Println("Success CREATE database")

    _,err = db.Exec("USE CurrencyConv")
   if err != nil {
       panic(err)
   }

   _,err = db.Exec("CREATE TABLE Currency ( FromC char(3), ToC char(3) )")
   if err != nil {
       panic(err)
   }

	fmt.Println("Success")
}