package main

import (
	"errors"
	dontlog "log"
	"net"
	"os"
	"os/signal"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/subparlabs/bonjourno/inputs"
	"github.com/subparlabs/bonjourno/log"
	"github.com/subparlabs/bonjourno/service"
)

var (
	// Data Sources
	say  = kingpin.Arg("static text", "Create a share with this text").Strings()
	file = kingpin.Flag("file", "Read messages from this file, periodically updating").ExistingFile()

	// How to slice the data
	words = kingpin.Flag("words", "Go through whole text, instead of lines").Bool()

	// How to choose messages
	random = kingpin.Flag("random", "Randomize messages, instead of sequential").Bool()

	// Message options
	interval = kingpin.Flag("interval", "Update interval for multiple strings (not watch, for ex), like 1s or 5m").Short('i').Default("5m").Duration()
	prefix   = kingpin.Flag("prefix", "Prefix all messages with this string").String()

	host = kingpin.Flag("host", "Host to broadast for the service").String()
	port = kingpin.Flag("port", "Port to broadast for the service").Int()
)

func main() {
	kingpin.Parse()

	msgChan, err := buildStream()
	if err != nil || msgChan == nil {
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

	serv, err := service.New(*host, *port, msgChan)
	if err != nil || serv == nil {
		log.Panic("Failed to create service", "err", err)
	}

	// Watch for signal to clean up before we exit
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	log.Info("Shutting down")
	serv.Stop()
	log.Info("Stopped service")
}

func buildStream() (<-chan string, error) {
	var source inputs.DataSource
	var err error
	if *file != "" {
		log.Info("Reading messages from file", "file", *file)
		source, err = inputs.FileWatcher(*file)
		if err != nil {
			return nil, err
		}
	}
	if source != nil && len(*say) > 0 {
		return nil, errors.New("Can only specify one data source")
	} else if len(*say) > 0 {
		log.Info("Using a static message", "msg", *say)
		source, err = inputs.StaticText(strings.Join(*say, " "))
		if err != nil {
			return nil, err
		}
	}
	if source == nil {
		return nil, errors.New("No source of data specified")
	}

	var builder inputs.MessageBuilder
	if *words {
		log.Info("Iterating words")
		builder = inputs.WordGroups(source)
	} else {
		log.Info("Iterating lines")
		builder = inputs.Lines(source)
	}

	var chooser inputs.MessageChooser
	if *random {
		log.Info("Randomizing order")
		chooser = inputs.RandomMessageChooser(builder)
	} else {
		log.Info("Sequential order")
		chooser = inputs.SequentialMessageChooser(builder)
	}

	return inputs.RateLimit(*interval,
		inputs.LimitSize(40,
			inputs.Cleanup(
				inputs.Prefix(*prefix,
					chooser)))), nil
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
