package main

import (
	dontlog "log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/subparlabs/bonjourno/log"
	"github.com/subparlabs/bonjourno/service"
)

var (
	say   = kingpin.Flag("say", "Create a share with this text").String()
	file  = kingpin.Flag("file", "Rotate through lines in this file").ExistingFile()
	watch = kingpin.Flag("watch", "Periodically update with the first line of this file").ExistingFile()

	host     = kingpin.Flag("host", "Host to broadast for the service").String()
	port     = kingpin.Flag("port", "Port to broadast for the service").Int()
	interval = kingpin.Flag("interval", "Update interval, like 1m or 10s").Short('i').Default("5m").Duration()
)

func main() {
	kingpin.Parse()

	numMethods := 0
	if *say != "" {
		numMethods++
	}
	if *file != "" {
		numMethods++
	}
	if *watch != "" {
		numMethods++
	}

	if numMethods == 0 {
		kingpin.FatalUsage("Must specify some way of saying something")
	} else if numMethods > 1 {
		kingpin.FatalUsage("Can only specify one way of saying something")
	}

	// Disable regular logging, cuz bonjour logs an error that isn't, and it's
	// confusing.
	dontlog.SetOutput(NullWriter{})

	if *host == "" {
		*host = getLocalIP()
	}
	if *port == 0 {
		*port = 45897
	}

	serv := service.New(*host, *port)
	serv.Say(*say)

	// Watch for signal to clean up before we exit
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Periodically update
	ticker := time.Tick(*interval)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker:
				serv.Say(time.Now().String())
			case <-signals:
				log.Info("Shutting down")
				serv.Stop()
				log.Info("Stopped service")
				return
			}
		}
	}()

	wg.Wait()
}

func getLocalIP() string {
	bestSoFar := "127.0.0.1"

	if addresses, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addresses {
			switch addr := address.(type) {
			case *net.IPNet:
				if !addr.IP.IsLoopback() {
					if addr4 := addr.IP.To4(); addr4 != nil {
						return addr4.String()
					}
					bestSoFar = addr.IP.String()
				}
			case *net.IPAddr:
				if !addr.IP.IsLoopback() {
					if addr4 := addr.IP.To4(); addr4 != nil {
						return addr4.String()
					}
					bestSoFar = addr.IP.String()
				}
			}
		}
	}

	return bestSoFar
}

type NullWriter struct{}

func (w NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
