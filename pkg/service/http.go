package service

import (
	"io/ioutil"
	"net/http"
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
}

func CreateServer(opts *Options) *HttpServer {
	return &HttpServer{
		PayloadSize: opts.PayloadSize,
		uploadSize:  opts.UploadSize,
		cli: &RemoteClient{
			upstream:       opts.Upstream,
			inParallel:     opts.QueryInParallel,
			longConnection: opts.LongConn,
			timeout:        opts.Timeout,
		},
	}
}

type HttpServer struct {
	PayloadSize int
	uploadSize  int
	cli         *RemoteClient
}

func (s *HttpServer) Loop() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ioutil.ReadAll(r.Body)
			r.Body.Close()
		}

		resp := s.cli.Query(s.uploadSize, s.PayloadSize)
		if len(resp) > s.PayloadSize {
			panic("query payload size exceeds the limit")
		}

		w.WriteHeader(http.StatusOK)
		if s.PayloadSize > 0 {
			selfPlayloadSize := s.PayloadSize - len(resp)
			if selfPlayloadSize > 0 {
				resp = rands.HumanFriendlyEnglishString(selfPlayloadSize) + resp
			}

			if _, err := w.Write([]byte(resp)); err != nil {
				panic(err)
			}
		}
	})

	http.ListenAndServe(":80", nil)
}
