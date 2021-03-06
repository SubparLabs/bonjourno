package service

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/oleksandr/bonjour"

	"github.com/subparlabs/bonjourno/log"
)

type Service struct {
	host string
	port int
	addr *net.TCPAddr

	messages   <-chan string
	currentMsg string

	stop chan struct{}
	wg   sync.WaitGroup
}

func New(host string, port int, msgChan <-chan string) (*Service, error) {
	s := &Service{
		host: host,
		port: port,

		messages: msgChan,

		stop: make(chan struct{}),
	}

	if addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port)); err != nil {
		return nil, err
	} else {
		s.addr = addr
	}

	listener, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return nil, err
	}
	log.Info("Listening for TCP connections", "address", s.addr)

	s.wg.Add(1)
	go s.serveTCP(listener)

	s.wg.Add(1)
	go s.start()

	return s, nil
}

func (s *Service) Stop() {
	defer s.wg.Wait()

	log.Info("Stopping service")

	close(s.stop)

	// TCPListener might still be blocked, so open a conn to it ourselves
	// to free it up.
	if conn, err := net.DialTCP("tcp", nil, s.addr); err != nil {
		log.Error("Failed to close down TCP listener", "err", err)
	} else {
		conn.Close()
		log.Info("Shut down TCP listener")
	}
}

func (s *Service) serveTCP(listener net.Listener) {
	defer s.wg.Done()

	for {
		select {
		case <-s.stop:
			return
		default:
		}

		if conn, err := listener.Accept(); err != nil {
			log.Error("Failed to accept TCP conn", "err", err)
		} else {
			s.wg.Add(1)
			go func(conn net.Conn) {
				defer s.wg.Done()
				defer conn.Close()

				if s.currentMsg != "" {
					if _, err := conn.Write([]byte(s.currentMsg + "\n")); err != nil {
						log.Error("Failed to write to TCP conn", "conn", conn, "err", err)
					}
				}
			}(conn)
		}
	}
}

func (s *Service) start() {
	defer s.wg.Done()

	var bonj *bonjour.Server
	defer func(b **bonjour.Server) {
		s.stopBonjour(*b)
	}(&bonj)

	for {
		select {
		case msg := <-s.messages:
			if msg != s.currentMsg {
				s.currentMsg = msg

				s.stopBonjour(bonj)
				bonj = nil

				if msg != "" {
					log.Info("Registering service", "name", msg, "host", s.host, "port", s.port)
					var err error
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
