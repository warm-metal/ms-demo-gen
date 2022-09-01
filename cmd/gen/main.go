package main

import (
	"flag"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	"github.com/warm-metal/ms-demo-gen.git/pkg/manifest"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"
	"gonum.org/v1/gonum/graph/encoding/dot"

	rands "github.com/xyproto/randomstring"
)

var (
	numServices             = flag.Int("services", 10, "Number of services in the demo")
	maxCaller               = flag.Int("max-caller", 2, "Maximum number of callers for each service except the root service")
	maxCallee               = flag.Int("max-callee", 3, "Maximum number of callees for each service except leaf services")
	maxReplicas             = flag.Int("max-replicas", 1, "Maximum number of replicas")
	longestCallChain        = flag.Int("longest-call-chain", -1, "Number of services in the longest call chain. -1 means not limit")
	outputDir               = flag.String("out", "", "The directory to where manifests to be generated. Manifests will be printed in the console if not specified.")
	targetNamespaces        = flag.String("namespaces", "", "Namespaces where workloads to be deployed. Multiple namespaces should be seperated by a comma(,).")
	image                   = flag.String("image", "docker.io/warmmetal/ms-demo-service:latest", "Image for each workload")
	alsoOutputTopology      = flag.Bool("gen-topology", true, "Output the topology in a DOT file.")
	payloadSize             = flag.Int("payload-size", 64, "The payload size of each backend")
	uploadSize              = flag.Int("upload-size", 0, "The uploaded data size of each request. If it is greater than 0, POST requests are issued, otherwize, GET requests instead. ")
	clientSizeTimeout       = flag.Duration("timeout", 0, "Client side timeout in time.Duration. 0 means never expire.")
	QueryInParallel         = flag.Bool("parallel", false, "If true, requests to all upstreams are issued at the same time. Otherwise, in the given order.")
	longConn                = flag.Bool("long", false, "If true, clients will use same L4 connection for precedure requests of the same upstream. Otherwise, build a new connection for each request.")
	numTrafficGenProc       = flag.Int("traffic-gen-proc", 1, "Number of concurrent processors per upstream")
	trafficGenQueryInterval = flag.Duration("traffic-gen-query-interval", time.Second, "Interval between queries of traffic generator.")
)

func main() {
	rands.Seed()
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	g := dag.New(&dag.Options{
		NumberVertices: *numServices,
		InDegreeRange:  [2]int{1, *maxCaller},
		OutDegreeRange: [2]int{0, *maxCallee},
		LongestWalk:    *longestCallChain,
	})

	if *alsoOutputTopology {
		dotBin, err := dot.Marshal(g, "", "", "")
		if err != nil {
			panic(err)
		}

		if err = ioutil.WriteFile(filepath.Join(*outputDir, "ws-demo.dot"), dotBin, 0755); err != nil {
			panic(err)
		}
	}

	var out io.WriteCloser
	if len(*outputDir) > 0 {
		out, err := os.Create(filepath.Join(*outputDir, "manifests.yaml"))
		if err != nil {
			panic(err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	manifest.GenForK8s(g, &manifest.Options{
		Options: service.Options{
			PayloadSize:     *payloadSize,
			UploadSize:      *uploadSize,
			Timeout:         *clientSizeTimeout,
			QueryInParallel: *QueryInParallel,
			LongConn:        *longConn,
		},
		TrafficGenOptions: manifest.TrafficGenOptions{
			NumConcurrentProc: *numTrafficGenProc,
			QueryInterval:     *trafficGenQueryInterval,
		},
		Output:             out,
		Namespaces:         strings.Split(*targetNamespaces, ","),
		ReplicaNumberRange: [2]int{1, *maxReplicas},
		Image:              *image,
	})
}
