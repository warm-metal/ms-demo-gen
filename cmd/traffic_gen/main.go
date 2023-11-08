package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/cmd/util"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"

	rands "github.com/xyproto/randomstring"
	_ "go.uber.org/automaxprocs"
)

func startQuery(ctx context.Context, c *service.RemoteClient, uploadReader *strings.Reader, discardPayload bool, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	exitted := false
	go func(exit *bool) {
		for {
			<-ctx.Done()
			*exit = true
		}
	}(&exitted)

	go func(done chan struct{}, exit *bool) {
		for {
			if *exit {
				break
			}

			if uploadReader != nil {
				uploadReader.Seek(0, io.SeekStart)
			}

			if discardPayload {
				c.Discard(nil, uploadReader)
			} else {
				fmt.Println(c.Query(nil, uploadReader, -1))
			}

			if interval > 0 {
				time.Sleep(interval)
			}
		}
		close(done)
	}(done, &exitted)

	return done
}

func main() {
	numProc := util.LookupEnv(util.ArgsKeyNumConcurrentProcess)
	queryInterval := util.LookupEnvDuration(util.ArgsKeyIntervalBetweenQueries)
	discardUpstreamPayload := util.LookupEnv(util.ArgsKeyDiscardUpstreamPayload)
	opts := &service.Options{
		UploadSize:      util.LookupEnv(util.ArgsKeyUploadSize),
		Upstream:        strings.Split(os.Getenv(util.ArgsKeyUpstream), ","),
		QueryInParallel: util.LookupEnv(util.ArgsKeyQueryInParallel) > 0,
		LongConn:        util.LookupEnv(util.ArgsKeyLongConnection) > 0,
		Timeout:         util.LookupEnvDuration(util.ArgsKeyQueryTimeout),
	}

	ctx := context.Background()
	waitingList := make([]<-chan struct{}, numProc)
	for i := 0; i < numProc; i++ {
		var uploadReader *strings.Reader
		if opts.UploadSize > 0 {
			uploadReader = strings.NewReader(rands.HumanFriendlyEnglishString(opts.UploadSize))
		}

		client := service.NewClient(opts)
		waitingList[i] = startQuery(ctx, &client, uploadReader, discardUpstreamPayload > 0, queryInterval)
	}

	for i := range waitingList {
		<-waitingList[i]
	}
}
