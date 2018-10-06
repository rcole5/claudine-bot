package main

import (
	"fmt"
	bolt "github.com/etcd-io/bbolt"
	"github.com/go-kit/kit/log"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/rcole5/claudine-bot"
	"github.com/rcole5/claudine-bot/bot"
	"net/http"
	"os"
	"os/signal"
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

	// Open up the db
	db, err := bolt.Open("my2.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var s claudine_bot.Service
	{
		s = claudine_bot.NewClaudineService(db)
	}

	var h http.Handler
	{
		h = claudine_bot.MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))
	}

	go bot.New(s, os.Getenv("USERNAME"), os.Getenv("TOKEN"), db)

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
