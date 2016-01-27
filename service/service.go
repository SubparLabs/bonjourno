package service

import (
	"regexp"
	"sync"
	"time"

	"github.com/oleksandr/bonjour"

	"github.com/subparlabs/bonjourno/log"
)

var (
	endsRe   = regexp.MustCompile("^[^a-zA-Z0-9-_]+|[^a-zA-Z0-9-_]+$")
	middleRe = regexp.MustCompile("[^a-zA-Z0-9-_]+")
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
	// Some characters cause the service to be ignored completely. Not sure
	// which, so make a conservative conversion.
	// TODO: look up the spec and only replace actually invalid chars

	// Just remove stuff at the start & end. This also serves to trim
	msg = endsRe.ReplaceAllString(msg, "")

	// Replace multiple invalid chars in middle with a single -
	msg = middleRe.ReplaceAllString(msg, "-")

	// The Finder sidebar cuts off somewhere under 20, maybe less, but
	// browsing to the share in "Network" shows somewhere around 40.
	if len(msg) > 40 {
		msg = msg[:40]
	}

	s.messages <- msg
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
	defer func(b **bonjour.Server) {
		s.stopBonjour(*b)
	}(&bonj)

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
			log.Info("Ending bonjour-updating routine")
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

		log.Info("Shutting down bonjour service")
		bonj.Shutdown()

		// I guess bonjour wants us to wait some unspecied
		// amount? This is what blocking or channels are for :/
		waitTime := time.Second * 5
		log.Info("Waiting for bonjour service to clean itself up", "waitTime", waitTime)
		time.Sleep(waitTime)
	}()
}
