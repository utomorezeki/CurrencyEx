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

func sevenMR(fromIn string, toIn string, db *sql.DB) (float32,float32,*map[string]float32){
	var (
		sevenDR map[string]float32 = make(map[string]float32)
		avg float32 = 0
		max float32 = 0
		min float32 = math.MaxFloat32
	)
	dataQuery := fmt.Sprintf("SELECT * FROM ExcData WHERE FromC='%s' AND ToC='%s'", fromIn,toIn)
	data, error := db.Query(dataQuery)
	if error != nil {
		log.Fatal(error)
	}
	defer data.Close()
	for data.Next() && len(sevenDR) < 7 {
		data2, error := db.Query(dataQuery)
		if error != nil {
			log.Fatal(error)
		}
		defer data2.Close()
		var maxRate float32
		maxDate, _ := time.Parse(shortForm, "0001-01-01")
		for data2.Next() {
			var (
				from   string
				to string
				date string
				rate float32
			)
			if err := data2.Scan(&from, &to,&date, &rate); err != nil {
				log.Fatal(err)
			}
			currDate, _ := time.Parse(shortForm, date) 
			if _, exist := sevenDR[currDate.String()[:10]];currDate.After(maxDate) && !exist{
				maxDate = currDate
				maxRate = rate
			}
		}
		sevenDR[maxDate.String()[:10]] = maxRate
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
	var placeHold bytes.Buffer
	for data.Next() {
		var (
			from string
			to string
		)
		if err := data.Scan(&from, &to); err != nil {
			log.Fatal(err)
		}
		placeHold.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", from,to))
	}
	page := &PageWriter{Data:template.HTML(placeHold.String())}

	from := r.FormValue("from")
	to := r.FormValue("to")
	p := &AddCurrPage{fromIn: from, toIn: to}
	if p.save() {
		page.Title = fmt.Sprintf("Successfully added Currency %s -> %s",from,to)
	} else {
		page.Title = "Fail to add! Currency Exists in the DB"
	}
	t, _ := template.ParseFiles("addCurrReact.html")
    t.Execute(w, page)
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

	_, err = time.Parse(shortForm, date)
	if err != nil{
		fmt.Fprintf(w, fmt.Sprintf("error", err))
		return
	} 
	p := &AddCurrPage{from,to,date,float32(rate)}
	p.save()
	t, _ := template.ParseFiles("dailyExReact.html")
    t.Execute(w, nil)
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
	from := r.FormValue("from")
	to := r.FormValue("to")

	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()
	avg,varia,sevenDR := sevenMR(from,to,db)

	var placeHold bytes.Buffer
	for dateKey, rateVal := range *sevenDR {
		placeHold.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%f</td></tr>", dateKey,rateVal))
	}
	page := &PageWriter{Title: fmt.Sprintf("From %s --> To %s",from,to), Data:template.HTML(placeHold.String()),Average: avg,Variance:varia}
	t, _ := template.ParseFiles("sevenDaysReact.html")
	t.Execute(w, page)
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
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
	var placeHold bytes.Buffer
	for data.Next() {
		var (
			from string
			to string
		)
		if err := data.Scan(&from, &to); err != nil {
			log.Fatal(err)
		}
		placeHold.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", from,to))
	}
	page := &PageWriter{Data:template.HTML(placeHold.String())}
	t, _ := template.ParseFiles("stop.html")
    t.Execute(w, page)
}

func stopReactHandler(w http.ResponseWriter, r *http.Request) {
	from := r.FormValue("from")
	to := r.FormValue("to")
	
	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()
	stringQuery := fmt.Sprintf("DELETE FROM Currency WHERE (FromC='%s' OR FromC='%s') AND (ToC='%s' OR ToC='%s')",from,to,from,to)
	data, error := db.Query(stringQuery)
	if error != nil {
		log.Fatal(error)
	}
	stringQuery = fmt.Sprintf("DELETE FROM ExcData WHERE (FromC='%s' OR FromC='%s') AND (ToC='%s' OR ToC='%s')",from,to,from,to)
	data, error = db.Query(stringQuery)
	if error != nil {
		log.Fatal(error)
	}
	defer data.Close()
	http.Redirect(w,r,"/stop/form", http.StatusFound)
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
	http.HandleFunc("/stop/form", stopHandler)
	http.HandleFunc("/stop/react", stopReactHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}