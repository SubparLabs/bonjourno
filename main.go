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

	var err error
	var inputStream service.InputStream
	if *say != "" {
		inputStream, err = service.NewStaticText(*say)
	} else if *file != "" {
		inputStream, err = service.NewFileLines(*file)
	} else if *watch != "" {
		inputStream, err = service.NewFileWatcher(*watch)
	}
	if err != nil {
		log.Error("Failed to create input stream", "err", err)
		os.Exit(1)
	} else if inputStream == nil {
		kingpin.FatalUsage("Need to specify something to say")
	}

	// Disable regular logging, cuz bonjour logs an error that isn't, and it's
	// confusing.
	dontlog.SetOutput(NullWriter{})

	if *host == "" {
		*host = getLocalIP()
	}
	if *port == 0 {
		// TODO: run a server that sends the same text on TCP conn
		*port = 45897
	}

	// Create a service, and init with something to say
	serv := service.New(*host, *port)
	serv.Say(inputStream.Get())

	// Watch for signal to clean up before we exit
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Periodically update
	log.Info("Checking for changes at", "interval", *interval)
	ticker := time.Tick(*interval)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker:
				serv.Say(inputStream.Get())
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
