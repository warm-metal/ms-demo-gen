package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/service"
)

func LookupEnv[TargetType int | time.Duration](key string) (t TargetType) {
	v := os.Getenv(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return TargetType(i)
}

const (
	argsKeyPayloadSize     = "ENV_PAYLOAD_SIZE"
	argsKeyUploadSize      = "ENV_UPLOAD_SIZE"
	argsKeyUpstream        = "ENV_UPSTREAM"
	argsKeyQueryInParallel = "ENV_QUERY_IN_PARALLEL"
	argsKeyLongConnection  = "ENV_USE_LONG_CONNECTION"
	argsKeyQueryTimeout    = "ENV_TIMEOUT"
)

func main() {
	timeoutDuration, err := time.ParseDuration(os.Getenv(argsKeyQueryTimeout))
	if err != nil {
		panic(err)
	}

	s := service.CreateServer(&service.Options{
		PayloadSize:     LookupEnv[int](argsKeyPayloadSize),
		UploadSize:      LookupEnv[int](argsKeyUploadSize),
		Upstream:        strings.Split(os.Getenv(argsKeyUpstream), ","),
		QueryInParallel: LookupEnv[int](argsKeyQueryInParallel) > 0,
		LongConn:        LookupEnv[int](argsKeyLongConnection) > 0,
		Timeout:         timeoutDuration,
		Address:         ":80",
	})

	ctx := context.Background()
	done := s.LoopInBackground(ctx)
	<-done
}
