package main

import (
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"time"
	"fmt"
	"math/rand"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnectUser struct {
	Websocket *websocket.Conn
	ClientIP string
}

type Product struct{
    ID int
    Model string
    Company string
    Price int
}

type Main struct{
    ID int
    EN string
}

type Entry struct {
  Name string
  Done bool
}

func newConnectUser(ws *websocket.Conn, clientIP string) *ConnectUser  {
	return &ConnectUser{
		Websocket: ws,
		ClientIP: clientIP,
	}
}


func IndexHandler(w http.ResponseWriter, r *http.Request)  {
	tmpl, _ := template.ParseFiles("templates/index.html")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var users = make(map[ConnectUser]int)

var sessions = map[string]string{}

func main() {
	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", content)

	http.HandleFunc("/add", addDataToDB)

	//goLogin
	http.HandleFunc("/login", login)

	//logOut
	http.HandleFunc("/logout", logout)

	http.HandleFunc("/ws", WebsocketHandler)

	log.Println("Listening on :80...")
	
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func content(w http.ResponseWriter, r *http.Request) {

	/*
	db, err := sql.Open("mysql", "root:@/productdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
    //rows, err := db.Query("select * from productdb.Products where price > ?", 70000)
    rows, err := db.Query("select * from productdb.Products")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    products := []product{}
     
    for rows.Next(){
        p := product{}
        err := rows.Scan(&p.id, &p.model, &p.company, &p.price)
        if err != nil{
            fmt.Println(err)
            continue
        }
        products = append(products, p)
    }
    */

	userSessionHash := r.URL.Path

	sessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		//fmt.Fprint(w, "Welcome, "+sessionID.)
		log.Println(sessionID)
		//w.Write([]byte())
		//fmt.Fprint(w, "Welcome http.ErrNoCookie, ")
		//return
		http.Redirect(w,r,"/login",http.StatusFound)
	}else if err != nil {
		fmt.Fprint(w, "Welcome PanicOnErr ")
		//PanicOnErr(err)
	}

	pingID, _ := r.Cookie("ping_id")


	//if len(sessions) > 0 { 
		userSessionHash = pingID.Value
	//}
	  

    t, err := template.ParseFiles("templates/content.html")
    if err != nil {
        fmt.Println(err)
    }

	entrys := []Entry{
			Entry{
				Name: "dfgdfgf",
				Done:  true,
			},
			Entry{
				Name: "dfgfdfdg",
				Done:   false,
			},
		}


	Langs := []string{"de","ru","ua"}	

	mains := getMainFromDB()


	
	//log.Println(products)	
	/*
    data := struct {
        Name string
        City string
        Products []Product
        List []Entry
    }{
        Name: userSessionHash,
        City: "MyCity",
        Products: products,
        List: entrys,
    }



      type a struct {
        Title   []string
        Article [][]string
      }

      var data = &a{
        Title: []string{"One", "Two", "Three"},
        Article: [][]string{
          []string{"a", "b", "c"},
          []string{"d", "e"},
          []string{"f", "g", "h", "i"}},
      }

      var tmplSrc = `
      {{range $i, $a := .Title}}
        Title: {{$a}}
        {{range $article := index $.Article $i}}
          Article: {{$article}}.
        {{end}}
      {{end}}`



    */

    data := struct {
        Name string
        City string
        Mains []Main
        List []Entry
        Langs []string
        Article [][]string
    }{
        Name: userSessionHash,
        City: "MyCity",
        Mains: mains,
        List: entrys,
        Langs: Langs,
        Article: [][]string{
          []string{"a", "b", "c"},
          []string{"d", "e"}},
    }

    //fmt.Printf("%v", data)

    t.Execute(w, data)
	
}

func login(w http.ResponseWriter, r *http.Request) {
    
	info := "Please sign in"

	if r.Method == "POST" {

		adminEmail := goDotEnvVariable("ADMIN_EMAIL")

		r.ParseForm()
	    inputLogin := r.Form["email"][0]

		if adminEmail == inputLogin {
			expiration := time.Now().Add(365 * 24 * time.Hour)

		    sessionID := RandStringRunes(32)
		    sessions[sessionID] = inputLogin

		    cookie := http.Cookie {
		    	Name: "session_id",
		    	Value: sessionID,
		    	Expires: expiration,
		    	HttpOnly: true,
		    }
		    http.SetCookie(w, &cookie)

		    log.Println("New post from login page cookie: "+sessionID)

		    pingCookie := http.Cookie {
		    	Name: "ping_id",
		    	Value: sessionID,
		    	Expires: expiration,
		    	HttpOnly: true,
		    }
		    http.SetCookie(w, &pingCookie)

		    http.Redirect(w,r,"/",http.StatusFound)
		} else {
			info = "Login or password incorrect!"
		}
	}

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	data := struct {
        Info string
    }{
        Info: info,
    }

	tmpl.Execute(w, data)
}


func logout(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
    	cookie := &http.Cookie {
		    Name:     "session_id",
		    Value:    "",
		    Path:     "/",
		    Expires: time.Unix(0, 0),
		    HttpOnly: true,
		}

		http.SetCookie(w, cookie)
		http.Redirect(w,r,"/login",http.StatusFound)
    }    
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request)  {
	ws, _ := upgrader.Upgrade(w, r, nil)

	defer func() {
		if err := ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	log.Println("Client connected:", ws.RemoteAddr().String())
	var socketClient *ConnectUser = newConnectUser(ws, ws.RemoteAddr().String())
	users[*socketClient] = 0
	log.Println("Number client connected ...", len(users))

	for {
		messageType, message, err := ws.ReadMessage()
		if  err != nil {
			log.Println("Ws disconnect waiting", err.Error())
			delete(users, *socketClient)
			log.Println("Number of client still connected ...", len(users))
			return
		}

		for client := range users {
			if err = client.Websocket.WriteMessage(messageType, message); err != nil {
				log.Println("Cloud not send Message to ", client.ClientIP, err.Error())
			}
		}

	}
}


func addDataToDB(w http.ResponseWriter, r *http.Request){
	db, err := sql.Open("mysql", "root:@/langsdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
     
    result, err := db.Exec("insert into langsdb.Products (model, company, price) values (?, ?, ?)", 
        RandStringRunes(20), RandStringRunes(10), rand.Intn(5000 - 1) + 1)
    if err != nil{
        panic(err)
    }
    fmt.Println(result.LastInsertId())  // id добавленного объекта
    fmt.Println(result.RowsAffected())  // количество затронутых строк
    http.Redirect(w,r,"/",http.StatusFound)
}

func getDataFromDB() []Product {
	db, err := sql.Open("mysql", "root:@/langsdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
    //rows, err := db.Query("select * from productdb.Products where price > ?", 70000)
    rows, err := db.Query("select * from langsdb.Products")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    products := []Product{}
     
    for rows.Next(){
        p := Product{}
        err := rows.Scan(&p.ID, &p.Model, &p.Company, &p.Price)
        if err != nil{
            fmt.Println(err)
            continue
        }
        products = append(products, p)
    }
    //for _, p := range products{
    //    fmt.Println(p.id, p.model, p.company, p.price)
    //}
    return products

    /* only one data

    row := db.QueryRow("select * from productdb.Products where id = ?", 2)
	prod := product{}
	err = row.Scan(&prod.id, &prod.model, &prod.company, &prod.price)
	if err != nil{
	    panic(err)
	}
	fmt.Println(prod.id, prod.model, prod.company, prod.price)

    */
}

func getMainFromDB() []Main {
	db, err := sql.Open("mysql", "root:@/langsdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
    //rows, err := db.Query("select * from productdb.Products where price > ?", 70000)
    rows, err := db.Query("select * from langsdb.main")
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    main := []Main{}
     
    for rows.Next(){
        m := Main{}
        err := rows.Scan(&m.ID, &m.EN)
        if err != nil{
            fmt.Println(err)
            continue
        }
        main = append(main, m)
    }
    //for _, p := range products{
    //    fmt.Println(p.id, p.model, p.company, p.price)
    //}
    return main

    /* only one data

    row := db.QueryRow("select * from productdb.Products where id = ?", 2)
	prod := product{}
	err = row.Scan(&prod.id, &prod.model, &prod.company, &prod.price)
	if err != nil{
	    panic(err)
	}
	fmt.Println(prod.id, prod.model, prod.company, prod.price)

    */
}

func updateDataDB(index int, data Product) {
	db, err := sql.Open("mysql", "root:@/langsdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
    result, err := db.Exec("update langsdb.Products set price = ? where id = ?", data.Price, index)
    if err != nil{
        panic(err)
    }
    fmt.Println(result.LastInsertId())
    fmt.Println(result.RowsAffected())
}

func deleteDataFromDB(index int) {
	db, err := sql.Open("mysql", "root:@/langsdb")
     
    if err != nil {
        panic(err)
    } 
    defer db.Close()
    result, err := db.Exec("delete from langsdb.Products where id = 1")
    if err != nil{
        panic(err)
    }
    fmt.Println(result.LastInsertId())  // id последнего удаленого объекта
    fmt.Println(result.RowsAffected())  // количество затронутых строк
}