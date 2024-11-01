package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	PORT string = ":8080"
	CONN string = "postgres://postgres:p4ssw0rd@localhost:5432/postgres?sslmode=disable"
)

func main() {
	db, err := sql.Open("postgres", CONN)
	if err != nil {
		log.Fatalf("error open database: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("could not reach database: %s", err)
	}

	gameRepository := NewGameRepository(db)
	gameService := NewGameService(gameRepository)
	gameServer := NewWebSocketServer(gameService)

	http.HandleFunc("/", serveHome)
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

	gameServer.Run()

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/index.html")
}
