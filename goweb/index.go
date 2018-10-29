package main

import (
	"log"
	"net/http"
	"html/template"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"strconv"
	"time"
	"bytes"
	"math"
)

const shortForm = "2006-01-02"

type AddCurrPage struct {
	fromIn  string
	toIn string
	dateIn  string
	rateIn float32
}

type PageWriter struct {
	Title  string
	Data template.HTML
	Average float32
	Variance float32
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
	if p.dateIn != "" {
		insQuery := fmt.Sprintf("INSERT INTO ExcData VALUES('%s','%s','%s','%f');", p.fromIn,p.toIn,p.dateIn,p.rateIn)
		insert, error := db.Query(insQuery)
		if error != nil {
			log.Fatal(error)
		}
		defer insert.Close()
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

func sevenDay(fromIn string, toIn string, timeIn time.Time, db *sql.DB) (float32,float32,*map[string]float32){
	var (
		sevenDR map[string]float32 = make(map[string]float32)
		avg float32 = 0
		max float32 = 0
		min float32 = math.MaxFloat32
	)
	for i := -3; i < 4; i++ {
		timeCur := timeIn.AddDate(0,0,i).String()[:10]
		dataQuery := fmt.Sprintf("SELECT * FROM ExcData WHERE Date='%s' AND FromC='%s' AND ToC='%s'",timeCur, fromIn,toIn)
		data, error := db.Query(dataQuery)
		if error != nil {
			log.Fatal(error)
		}
		defer data.Close()
		for data.Next() {
			var (
				from   string
				to string
				date string
				rate float32
			)
			if err := data.Scan(&from, &to,&date, &rate); err != nil {
				log.Fatal(err)
			}
			sevenDR[date] = rate
		}
	}
	for _, rateVal := range sevenDR {
		avg = avg + rateVal
		if rateVal > max { 
			max = rateVal
		}
		if rateVal < min {
			min = rateVal
		}
	}
	avg = avg / float32(len(sevenDR))
	return avg,max - min,&sevenDR
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
}

func dailyExHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("dailyEx.html")
    t.Execute(w, nil)
}

func dailyExReactHandler(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	from := r.FormValue("from")
	to := r.FormValue("to")

	rate, err := strconv.ParseFloat(r.FormValue("rate"), 32)
	if err != nil {
		fmt.Fprintf(w, "rate must be numerical")
		return
	}
	p := &AddCurrPage{from,to,date,float32(rate)}
	p.save()
}

func dateShowHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("dateShow.html")
    t.Execute(w, nil)
}

func dateShowReactHandler(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	parsedTime, err := time.Parse(shortForm, date)
	if err != nil{
		fmt.Fprintf(w, fmt.Sprintf("error", err))
		return
	} 

	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()

	dataQuery := fmt.Sprintf("SELECT * FROM ExcData WHERE Date='%s'",date)
	data, error := db.Query(dataQuery)
	if error != nil {
		log.Fatal(error)
	}
	defer data.Close()
	var placeHold bytes.Buffer
	for data.Next() {
		var (
			from   string
			to string
			date string
			rate float32
			avg float32
		)
		if err := data.Scan(&from, &to,&date, &rate); err != nil {
			log.Fatal(err)
		}
		avg,_,_ = sevenDay(from,to,parsedTime,db)
		placeHold.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%f</td><td>%f</td></tr>", from, to,rate,avg))
	}
	page := &PageWriter{Title: date, Data:template.HTML(placeHold.String())}
	
	t, _ := template.ParseFiles("dateShowReact.html")
	t.Execute(w, page)
	
}

func sevenDaysHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("sevenDays.html")
    t.Execute(w, nil)
}

func sevenDaysReactHandler(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	from := r.FormValue("from")
	to := r.FormValue("to")
	parsedTime, err := time.Parse(shortForm, date)
	if err != nil{
		fmt.Fprintf(w, fmt.Sprintf("error", err))
		return
	} 

	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()
	avg,varia,sevenDR := sevenDay(from,to, parsedTime,db)

	var placeHold bytes.Buffer
	for dateKey, rateVal := range *sevenDR {
		placeHold.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%f</td></tr>", dateKey,rateVal))
	}
	page := &PageWriter{Title: fmt.Sprintf("From %s --> To %s",from,to), Data:template.HTML(placeHold.String()),Average: avg,Variance:varia}
	t, _ := template.ParseFiles("sevenDaysReact.html")
	t.Execute(w, page)
}

func main() {
	http.HandleFunc("/addCurr/form", addCurrHandler)
	http.HandleFunc("/addCurr/react", addCurrReactHandler)
	http.HandleFunc("/dailyEx/form", dailyExHandler)
	http.HandleFunc("/dailyEx/react", dailyExReactHandler)
	http.HandleFunc("/dateShow/form", dateShowHandler)
	http.HandleFunc("/dateShow/react", dateShowReactHandler)
	http.HandleFunc("/sevenDays/form", sevenDaysHandler)
	http.HandleFunc("/sevenDays/react", sevenDaysReactHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}