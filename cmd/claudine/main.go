package main

import (
	"flag"
	"fmt"
	"github.com/rcole5/claudine-bot"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var s claudine_bot.Service
	{
		// TODO: Pass in db connection
		s = claudine_bot.NewClaudineService()
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	log.Fatal("exit", <-errs)
}
