package main

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/cmd/util"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"

	rands "github.com/xyproto/randomstring"
)

func main() {
	rands.Seed()
	rand.Seed(time.Now().UnixNano())

	s := service.CreateServer(&service.Options{
		PayloadSize:     util.LookupEnv(util.ArgsKeyPayloadSize),
		UploadSize:      util.LookupEnv(util.ArgsKeyUploadSize),
		Upstream:        strings.Split(os.Getenv(util.ArgsKeyUpstream), ","),
		QueryInParallel: util.LookupEnv(util.ArgsKeyQueryInParallel) > 0,
		LongConn:        util.LookupEnv(util.ArgsKeyLongConnection) > 0,
		Timeout:         util.LookupEnvDuration(util.ArgsKeyQueryTimeout),
		Address:         ":80",
	})

	ctx := context.Background()
	done := s.LoopInBackground(ctx)
	<-done
}
