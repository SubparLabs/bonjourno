package service

import (
	"regexp"
	"sync"
	"time"

	"github.com/oleksandr/bonjour"

	"github.com/subparlabs/bonjourno/log"
)

var (
	nameRe = regexp.MustCompile("[^a-zA-Z0-9-_ ]")
)

type Service struct {
	host string
	port int

	messages chan string
	stop     chan struct{}
	wg       sync.WaitGroup
}

func New(host string, port int) *Service {
	s := &Service{
		host: host,
		port: port,

		messages: make(chan string),
		stop:     make(chan struct{}),
	}

	s.wg.Add(1)
	go s.start()

	return s
}

func (s *Service) Say(msg string) {
	s.messages <- nameRe.ReplaceAllString(msg, "-")
}

func (s *Service) Stop() {
	defer s.wg.Wait()

	log.Info("Stopping service")
	close(s.stop)
	s.stop = nil
}

func (s *Service) start() {
	defer s.wg.Done()

	var bonj *bonjour.Server
	defer s.stopBonjour(bonj)

	var err error
	var msg string

	for {
		select {
		case newMsg := <-s.messages:
			if newMsg != msg {
				msg = newMsg

				s.stopBonjour(bonj)

				log.Info("Registering service", "name", msg, "host", s.host, "port", s.port)
				bonj, err = bonjour.RegisterProxy(
					msg,
					"_afpovertcp._tcp", "local",
					s.port, s.host, s.host,
					nil, nil)
				if err != nil || bonj == nil {
					log.Error("Failed to register service with bonjour", "err", err)
					bonj = nil
				}
			}
		case <-s.stop:
			return
		}
	}
}

func (s *Service) stopBonjour(bonj *bonjour.Server) {
	if bonj == nil {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		bonj.Shutdown()

		// I guess bonjour wants us to wait some unspecied
		// amount? This is what blocking or channels are for :/
		time.Sleep(3 * time.Second)
	}()
}
