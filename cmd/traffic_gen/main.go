package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/cmd/util"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"

	rands "github.com/xyproto/randomstring"
)

func startQuery(ctx context.Context, c *service.RemoteClient, uploadData string, interval time.Duration) <-chan struct{} {
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

			c.Query(strings.NewReader(uploadData), -1)
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
		waitingList[i] = startQuery(ctx, service.NewClient(opts), rands.HumanFriendlyEnglishString(opts.UploadSize), queryInterval)
	}

	for i := range waitingList {
		<-waitingList[i]
	}
}
