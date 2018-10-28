package main

import (
	"log"
	"net/http"
	"html/template"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

type AddCurrPage struct {
	fromIn  string
	toIn string
	dateIn  string
	rateIn string
}

func (p *AddCurrPage) save() bool {
	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()
	data, error := db.Query("SELECT * FROM Currency")
	if error != nil {
		log.Fatal(error)
	}
	defer data.Close()
	var exist bool = false
	for (data.Next() && !exist) {
		var (
			from string
			to string
		)
		if err := data.Scan(&from, &to); err != nil {
			log.Fatal(err)
		}
		exist = p.checkExist(from,to)
	}
	if exist {
		return false;
	} else {
		insQuery := fmt.Sprintf("INSERT INTO Currency VALUES('%s','%s');", p.fromIn,p.toIn)
		insert, error := db.Query(insQuery)
		if error != nil {
			log.Fatal(error)
		}
		defer insert.Close()
		return true;
	}
}

func (p *AddCurrPage) checkExist(dbFrom string, dbTo string) bool {
	if p.fromIn == p.toIn {
		return false
	}
	checkFrom := ((p.fromIn == dbFrom) || (p.fromIn == dbTo))
	checkTo := ((p.toIn == dbFrom) || (p.toIn == dbTo))
	return (checkFrom && checkTo)
}

func show() {
	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()

	data, error := db.Query("SELECT * FROM Currency")
	if error != nil {
		log.Fatal(error)
	}
	defer data.Close()
	for data.Next() {
		var (
			id   string
			name string
		)
		if err := data.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("id %s name is %s\n", id, name)
	}
}

func addCurrHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("addCurr.html")
    t.Execute(w, nil)
}

func addCurrReactHandler(w http.ResponseWriter, r *http.Request) {
	from := r.FormValue("from")
	to := r.FormValue("to")
	p := &AddCurrPage{fromIn: from, toIn: to}
	if p.save() {
		fmt.Fprintf(w, "Success currency input")
	} else {
		fmt.Fprintf(w, "Fail to input! currency already exist")
	}
	show()
}

func dailyExHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("dailyEx.html")
    t.Execute(w, nil)
}

func dailyExReactHandler(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	from := r.FormValue("from")
	to := r.FormValue("to")
	rate := r.FormValue("rate")
	http.Redirect(w, r, "/dailyEx/form", http.StatusFound)
	fmt.Printf("Date : %s\nFrom : %s\nTo : %s\nRate : %s\n",date, from, to,rate)
}

func main() {
	http.HandleFunc("/addCurr/form", addCurrHandler)
	http.HandleFunc("/addCurr/react", addCurrReactHandler)
	http.HandleFunc("/dailyEx/form", dailyExHandler)
	http.HandleFunc("/dailyEx/react", dailyExReactHandler)
	//http.HandleFunc("/edit/", editHandler)
    //http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}