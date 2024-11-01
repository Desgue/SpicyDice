package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Desgue/SpicyDice/internal/repository"
	"github.com/Desgue/SpicyDice/internal/server"
	"github.com/Desgue/SpicyDice/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("error open database: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("could not reach database: %s", err)
	}

	gameRepository := repository.NewGameRepository(db)
	gameService := service.NewGameService(gameRepository)
	gameServer := server.NewWebSocketServer(gameService)

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
