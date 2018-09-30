package main

import (
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/joho/godotenv"
	"github.com/rcole5/claudine-bot"
	"github.com/rcole5/claudine-bot/bot"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	// Load the settings
	_ := godotenv.Load();

	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var s claudine_bot.Service
	{
		// TODO: Pass in db connection
		s = claudine_bot.NewClaudineService()
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
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
