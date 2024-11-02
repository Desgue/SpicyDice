package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Desgue/SpicyDice/internal/config"
	"github.com/Desgue/SpicyDice/internal/repository"
	"github.com/Desgue/SpicyDice/internal/server"
	"github.com/Desgue/SpicyDice/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	conf := config.New()
	connStr := conf.Postgres.String()

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
