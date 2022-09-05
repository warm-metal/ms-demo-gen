package service

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
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
		cli: NewClient(opts),
		server: &http.Server{
			Addr: opts.Address,
		},
	}

	if opts.UploadSize > 0 {
		s.uploadReader = strings.NewReader(rands.HumanFriendlyEnglishString(opts.UploadSize))
	}

	if opts.PayloadSize > 0 {
		s.payloadReader = strings.NewReader(rands.HumanFriendlyEnglishString(opts.PayloadSize))
	}

	s.serveMux.HandleFunc("/", s.root)
	s.server.Handler = &s.serveMux
	return s
}

type HttpServer struct {
	payloadReader *strings.Reader
	uploadReader  *strings.Reader
	cli           RemoteClient
	serveMux      http.ServeMux
	server        *http.Server
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

	if s.uploadReader != nil {
		s.uploadReader.Seek(0, io.SeekStart)
	}

	s.cli.Discard(s.uploadReader)
	w.WriteHeader(http.StatusOK)
	if s.payloadReader != nil {
		s.payloadReader.Seek(0, io.SeekStart)
		if _, err := io.Copy(w, s.payloadReader); err != nil {
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
			return
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
