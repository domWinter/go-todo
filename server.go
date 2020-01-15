package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-redis/redis/v7"
	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	Text string `json:"text"`
	Id   int    `json:"id"`
}

type Todo struct {
	Id    int
	Title string
	Done  bool
}

type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

func RedisNewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
	// Output: PONG <nil>
}

func main() {
	fmt.Println("Hello World")

	database, _ := sql.Open("sqlite3", "./entries.db")
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS entries (id INTEGER PRIMARY KEY, text TEXT, status INTEGER)")
	statement.Exec()

	//client := RedisNewClient()

	tmpl := template.Must(template.ParseFiles("layout.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//val, err := client.Get("Todo").Result()
		//if err != nil {
		//	panic(err)
		//}

		var data TodoPageData

		data.PageTitle = "TODO list"

		database, _ := sql.Open("sqlite3", "./entries.db")
		rows, _ := database.Query("SELECT id, text, status FROM entries")

		var id int
		var text string
		var status int
		for rows.Next() {
			rows.Scan(&id, &text, &status)
			var todo = Todo{Id: id, Title: text, Done: !(status == 0)}
			data.Todos = append(data.Todos, todo)
		}

		tmpl.Execute(w, data)
	})

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodDelete:
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			json_string := string(body)
			var e Entry
			json.Unmarshal([]byte(json_string), &e)

			statement, _ = database.Prepare("DELETE FROM entries WHERE id = ?")
			statement.Exec(e.Id)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(e)
			return
		case http.MethodPost:
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			json_string := string(body)

			var e Entry
			json.Unmarshal([]byte(json_string), &e)

			statement, _ = database.Prepare("INSERT INTO entries (text, status) VALUES (?, ?)")
			statement.Exec(e.Text, 0)

			//err = client.Set("Todo", e.Text, 0).Err()

			//if err != nil {
			//	panic(err)
			//}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(e)
			return
		default:
			http.NotFound(w, r)
			return
		}

	})
	log.Fatal(http.ListenAndServe(":7777", nil))
}
