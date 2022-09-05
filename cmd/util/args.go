package util

import (
	"os"
	"strconv"
	"time"
)

func LookupEnv[TargetType int](key string) (t TargetType) {
	v := os.Getenv(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(key)
	}
	return TargetType(i)
}

const (
	ArgsKeyPayloadSize            = "ENV_PAYLOAD_SIZE"
	ArgsKeyUploadSize             = "ENV_UPLOAD_SIZE"
	ArgsKeyUpstream               = "ENV_UPSTREAM"
	ArgsKeyQueryInParallel        = "ENV_QUERY_IN_PARALLEL"
	ArgsKeyLongConnection         = "ENV_USE_LONG_CONNECTION"
	ArgsKeyQueryTimeout           = "ENV_TIMEOUT"
	ArgsKeyNumConcurrentProcess   = "ENV_CONCURRENT_PROCS"
	ArgsKeyIntervalBetweenQueries = "ENV_INTERVAL_BETWEEN_QUERIES"
	ArgsKeyDiscardUpstreamPayload = "ENV_DISCARD_UPSTREAM_PAYLOAD"
)

func LookupEnvDuration(key string) time.Duration {
	duration, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		panic(err)
	}

	return duration
}
