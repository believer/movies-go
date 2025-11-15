package main

import (
	"believer/movies/utils"
	"context"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

func levelFromEnv(name string) log.Level {
	switch strings.ToLower(name) {
	case "trace":
		return log.LevelTrace
	case "debug":
		return log.LevelDebug
	case "info", "":
		return log.LevelInfo
	case "warn", "warning":
		return log.LevelWarn
	case "error":
		return log.LevelError
	case "fatal":
		return log.LevelFatal
	case "panic":
		return log.LevelPanic
	default:
		return log.LevelInfo
	}
}

func main() {
	log.SetLevel(levelFromEnv(os.Getenv("LOG_LEVEL")))
	err := utils.LoadEnv()
	ctx := context.Background()
	appEnv := os.Getenv("APP_ENV")

	if err != nil && appEnv == "development" {
		log.Error("Server could not start", err)
		os.Exit(1)
	}

	config := config{
		addr: ":8080",
		db: dbConfig{
			dsn: os.Getenv("DATABASE_URL"),
		},
	}

	// DB
	conn, err := pgx.Connect(ctx, config.db.dsn)

	if err != nil {
		panic(err)
	}

	log.Info("Connected to database")
	defer conn.Close(ctx)

	api := api{
		config: config,
		db:     conn,
	}

	port := os.Getenv("PORT")

	if port != "" {
		api.config.addr = port
	}

	if err := api.run(api.mount()); err != nil {
		log.Error("Server could not start", err)
		os.Exit(1)
	}
}
