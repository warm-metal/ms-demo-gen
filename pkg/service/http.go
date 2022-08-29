package service

import (
	"net/http"

	rands "github.com/xyproto/randomstring"
)

type Options struct {
	PayloadSize     int
	Upstream        []string
	QueryInParallel bool
	LongConn        bool

	// FIXME
	Timeout int
	UploadSize int
}

func CreateServer(opts *Options) *HttpServer {
	return &HttpServer{
		PayloadSize: opts.PayloadSize,
		cli: &RemoteClient{
			upstream: opts.Upstream,
			inParallel: opts.QueryInParallel,
			longConnection: opts.LongConn,
		},
	}
}

type HttpServer struct {
	PayloadSize int
	cli         *RemoteClient
}

func (s *HttpServer) Loop() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := s.cli.Query(s.PayloadSize)
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
