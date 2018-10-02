package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/rcole5/claudine-bot"
	"github.com/rcole5/claudine-bot/bot"
	"github.com/rcole5/claudine-bot/models"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	// Load the settings
	godotenv.Load()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	db, err := gorm.Open("sqlite3", "commands.db")
	if err != nil {
		panic(err)
	}

	// Migrate the db
	db.AutoMigrate(&models.Command{})

	var s claudine_bot.Service
	{
		s = claudine_bot.NewClaudineService(db)
	}

	var commands []*models.Command
	db.Find(&commands)
	for _, comm := range commands {
		s.NewCommand(context.Background(), comm.Channel, claudine_bot.Command{
			Trigger: comm.Trigger,
			Action: comm.Action,
		})
	}

	var h http.Handler
	{
		h = claudine_bot.MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))
	}

	go bot.New(s, os.Getenv("USERNAME"), os.Getenv("TOKEN"), strings.Split(os.Getenv("CHANNELS"), ","))

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", ":"+os.Getenv("PORT"))
		errs <- http.ListenAndServe(":"+os.Getenv("PORT"), h)
	}()

	logger.Log("exit", <-errs)
}
