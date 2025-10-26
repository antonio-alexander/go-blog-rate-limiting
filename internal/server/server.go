package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/antonio-alexander/go-blog-rate-limiting/internal/config"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/data"
	"github.com/antonio-alexander/go-blog-rate-limiting/internal/limiter"

	"github.com/pkg/errors"
)

const serverLogPrefix string = "[server] "

type server struct {
	sync.WaitGroup
	sync.RWMutex
	config struct {
		host           string
		port           string
		algorithm      string
		tokenReplinish time.Duration
	}
	rateLimiter limiter.Limiter
	httpServer  *http.Server
	chError     chan error
}

func New(parameters ...any) Server {
	s := &server{}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *config.Configuration:
			s.config.host = p.Host
			s.config.port = p.Port
			s.config.algorithm = p.Algorithm
			s.config.tokenReplinish = p.TokenReplinish
		case limiter.Limiter:
			s.rateLimiter = p
		}
	}
	return s
}

func (s *server) errorHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if _, err = w.Write([]byte(err.Error())); err != nil {
		fmt.Printf(serverLogPrefix+"error while writing bytes: %s\n", err.Error())
	}
}

func (s *server) endpointWait(w http.ResponseWriter, r *http.Request) {
	//get request
	request := data.NewRequest()
	if err := request.FromRequest(r); err != nil {
		s.errorHandler(w, err)
		return
	}

	//execute business logic
	select {
	case <-time.After(request.Wait):
		fmt.Printf(serverLogPrefix+"%s: wait of %v completed\n", request.Id, request.Wait)
	case <-r.Context().Done():
		fmt.Printf(serverLogPrefix+"%s: wait cancelled (context)\n", request.Id)
		return
	}

	//marshal response
	bytes, err := json.Marshal(&data.Response{
		Id:            request.Id,
		ApplicationId: request.ApplicationId,
		Wait:          request.Wait,
		Weight:        request.Weight,
	})
	if err != nil {
		s.errorHandler(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
	if _, err = w.Write(bytes); err != nil {
		fmt.Printf(serverLogPrefix+"error while writing bytes: %s\n", err.Error())
	}
}

func (s *server) launchServer() error {
	started := make(chan struct{})
	s.Add(1)
	go func() {
		defer s.Done()

		close(started)
		fmt.Printf(serverLogPrefix+"server listening on %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.chError <- err
			}
		}
	}()
	<-started
	select {
	case <-time.After(time.Second): //wait one second for error
		return nil
	case err := <-s.chError:
		return err
	}
}

func (s *server) Start() error {
	s.Lock()
	defer s.Unlock()

	mux := http.NewServeMux()
	mux.Handle(data.MethodWait+" "+data.RouteWait, s.rateLimiter.Middleware(s.endpointWait))
	httpServer := &http.Server{Handler: mux}
	httpServer.Addr = s.config.host
	if s.config.port != "" {
		httpServer.Addr = ":" + s.config.port
	}
	s.chError = make(chan error, 1)
	s.httpServer = httpServer
	return s.launchServer()
}

func (s *server) Stop() error {
	s.Lock()
	defer s.Unlock()

	defer func() {
		close(s.chError)
	}()
	if err := s.httpServer.Close(); err != nil {
		return err
	}
	s.Wait()
	select {
	case <-time.After(time.Second): //wait one second for error
		return nil
	case err := <-s.chError:
		return err
	}
}
