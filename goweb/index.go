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
)

const shortForm = "2006-01-02"

type AddCurrPage struct {
	fromIn  string
	toIn string
	dateIn  string
	rateIn float32
}

type PageWriter struct {
	title  string
	data []string
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

func sevenDay(fromIn string, toIn string, timeIn time.Time, db *sql.DB) (float32,*[]string,*[]float32){
	var (
		sevenD []string
		sevenR []float32
		avg float32 = 0
	)
	fmt.Println("INSIDE SEVENDAY")
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
			sevenD = append(sevenD, date)
			sevenR = append(sevenR, rate)
			fmt.Printf("DATE %s RATE %f \n", date,rate)
		}
	}
	for _, dataEl := range sevenR {
		avg = avg + dataEl
	}
	avg = avg / float32(len(sevenR))
	return avg,&sevenD,&sevenR
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
	//THIS LINE IS A PLACEHOLDER
	fmt.Printf(parsedTime.String()[:10])

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
	var placeHold []string
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
		placeHold = append(placeHold,fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%f</td><td></td></tr>", from, to,rate))
		avg,_,_ = sevenDay(from,to,parsedTime,db)
		fmt.Printf("FROM %s TO %s RATE %f AVG-7DAY %f \n", from, to,rate,avg)
	}
	page := &PageWriter{date, placeHold}
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
	//THIS LINE IS A PLACEHOLDER
	fmt.Printf(parsedTime.String()[:10])

	db, error := sql.Open("mysql","root:pass@tcp(sql1:3306)/CurrencyConv")
	if error != nil {
		panic(error)
	}	
	defer db.Close()
	sevenDay(from,to, parsedTime,db)
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
    //http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}