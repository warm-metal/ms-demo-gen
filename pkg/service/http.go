package service

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	rands "github.com/xyproto/randomstring"
)

type Options struct {
	PayloadSize     int
	UploadSize      int
	Timeout         time.Duration
	Upstream        []string
	QueryInParallel bool
	LongConn        bool
	Address         string
}

func CreateServer(opts *Options) *HttpServer {
	s := &HttpServer{
		PayloadSize: opts.PayloadSize,
		uploadSize:  opts.UploadSize,
		cli: &RemoteClient{
			upstream:       opts.Upstream,
			inParallel:     opts.QueryInParallel,
			longConnection: opts.LongConn,
			timeout:        opts.Timeout,
		},
		server: &http.Server{
			Addr: opts.Address,
		},
	}

	s.serveMux.HandleFunc("/", s.root)
	s.server.Handler = &s.serveMux
	return s
}

type HttpServer struct {
	PayloadSize int
	uploadSize  int
	cli         *RemoteClient
	serveMux    http.ServeMux
	server      *http.Server
}

func (s *HttpServer) root(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			panic(err)
		}
		io.Copy(devNull, r.Body)
		devNull.Close()
		r.Body.Close()
	}

	s.cli.Discard(s.uploadSize)
	w.WriteHeader(http.StatusOK)
	if s.PayloadSize > 0 {
		if _, err := w.Write([]byte(rands.HumanFriendlyEnglishString(s.PayloadSize))); err != nil {
			panic(err)
		}
	}
}

func (s *HttpServer) LoopInBackground(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			<-ctx.Done()
			s.server.Close()
		}
	}()

	go func(done chan struct{}) {
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
		close(done)
	}(done)

	return done
}
