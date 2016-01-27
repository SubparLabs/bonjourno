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

	inputStream, err := buildStream()
	if err != nil || inputStream == nil {
		kingpin.FatalUsage("Failed to create input stream: " + err.Error())
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

	serv, err := service.New(*host, *port)
	if err != nil || serv == nil {
		log.Panic("Failed to create service", "err", err)
	}

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

		// Init with something to say before waiting for the next interval
		serv.Say(inputStream.Get())

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

func buildStream() (service.InputStream, error) {
	var streams []service.InputStream

	if *watch != "" {
		if stream, err := service.NewFileWatcher(*watch); err != nil {
			return nil, err
		} else {
			streams = append(streams, stream)
		}
	}
	if *file != "" {
		if stream, err := service.NewFileLines(*file); err != nil {
			return nil, err
		} else {
			streams = append(streams, stream)
		}
	}
	if *say != "" {
		if stream, err := service.NewStaticText(*say); err != nil {
			return nil, err
		} else {
			streams = append(streams, stream)
		}
	}

	return service.NewPriorityMultistream(streams)
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
