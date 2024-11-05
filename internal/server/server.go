package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/dotenv-org/godotenvvault/autoload"

	"github.com/frsfahd/go-expenseTracker/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}

	// check status DB
	stats := NewServer.db.Health()
	if stats["status"] == "up" {
		slog.Info("database connected ✅")
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// check server status
	host, _ := os.Hostname()
	slog.Info("server up at ", host+":", os.Getenv("PORT"))

	return server
}
