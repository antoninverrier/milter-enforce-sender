package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d--j/go-milter/mailfilter"
)

var cfg MilterConfig
var users UserDb

func mainLoop() {

	milter, err := mailfilter.New(cfg.BindProto, cfg.BindHost, doMilter, mailfilter.WithoutBody())

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting (%s:%s)", milter.Addr().Network(), milter.Addr().String())

	if cfg.DrynRun {
		log.Printf("[!] dry run mode, will not actually rewrite messages")
	}

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)

	go func() {

		for {
			s := <-sig
			switch s {
			case syscall.SIGHUP:
				log.Printf("reloading (%s:%s)", milter.Addr().Network(), milter.Addr().String())
				reloadIdentities()
			default:
				log.Printf("exiting (%s:%s)", milter.Addr().Network(), milter.Addr().String())
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				milter.Shutdown(ctx)
				cancel()
			}
		}

	}()

	milter.Wait()
}

func main() {
	setupFlags()
	reloadIdentities()
	mainLoop()
}
