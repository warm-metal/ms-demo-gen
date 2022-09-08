package main

import (
	"flag"
	"fmt"
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
	"k8s.io/apimachinery/pkg/api/resource"

	rands "github.com/xyproto/randomstring"
)

var (
	numServices             = flag.Int("services", 10, "Number of services in the demo")
	maxCaller               = flag.Int("max-downstream", 2, "Maximum number of callers for each service except the root service")
	maxCallee               = flag.Int("max-upstream", 3, "Maximum number of callees for each service except leaf services")
	maxReplicas             = flag.Int("max-replicas", 1, "Maximum number of replicas")
	maxVersions             = flag.Int("max-versions", 1, "Maximum number of versions for each service")
	longestCallChain        = flag.Int("longest-call-chain", -1, "Number of services in the longest call chain. -1 means not limit")
	outputDir               = flag.String("out", "", "The directory to where manifests to be generated. Manifests will be printed in the console if not specified.")
	targetNamespaces        = flag.String("namespaces", "", "Namespaces where workloads to be deployed. Multiple namespaces should be seperated by a comma(,). You need to create those namespaces manually.")
	image                   = flag.String("image", "docker.io/warmmetal/ms-demo-service:latest", "Image for each workload")
	alsoOutputTopology      = flag.Bool("gen-topology", true, "Output the topology in a DOT file.")
	payloadSize             = flag.String("payload-size", "64", "The payload size of each backend. Such as 10Ki, 1Mi")
	uploadSize              = flag.String("upload-size", "0", "The uploaded data size of each request. Such as 1Ki. If it is greater than 0, POST requests are issued, otherwize, GET requests instead. ")
	clientSideTimeout       = flag.Duration("timeout", 0, "Client side timeout in time.Duration. 0 means never expire.")
	QueryInParallel         = flag.Bool("parallel", false, "If true, requests to all upstreams are issued at the same time. Otherwise, in the given order.")
	longConn                = flag.Bool("long", false, "If true, clients will use same L4 connection for precedure requests of the same upstream. Otherwise, build a new connection for each request.")
	numTrafficGenProc       = flag.Int("traffic-gen-proc", 1, "Number of concurrent processors per upstream")
	trafficGenQueryInterval = flag.Duration("traffic-gen-query-interval", time.Second, "Interval between queries of traffic generator.")
	cpuRequest              = flag.String("service-cpu-request", "", "CPU fragments requested for each service.")
	cpuLimit                = flag.String("service-cpu-limit", "", "CPU fragments limited for each service.")
)

func main() {
	rands.Seed()
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	g := dag.New(&dag.Options{
		NumberVertices:      *numServices,
		InDegreeRange:       [2]int{1, *maxCaller},
		OutDegreeRange:      [2]int{0, *maxCallee},
		NumberVersionsRange: [2]int{1, *maxVersions},
		LongestWalk:         *longestCallChain,
	})

	app := fmt.Sprintf("msd%d-%s", *numServices, rands.HumanFriendlyEnglishString(5))
	if *alsoOutputTopology {
		dotBin, err := dot.Marshal(g, app, "", "")
		if err != nil {
			panic(err)
		}

		if err = ioutil.WriteFile(filepath.Join(*outputDir, fmt.Sprintf("topology-%s.dot", app)), dotBin, 0755); err != nil {
			panic(err)
		}
	}

	var out io.WriteCloser
	if len(*outputDir) > 0 {
		out, err := os.Create(filepath.Join(*outputDir, fmt.Sprintf("manifests-%s.yaml", app)))
		if err != nil {
			panic(err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	payloadZ := resource.MustParse(*payloadSize)
	uploadZ := resource.MustParse(*uploadSize)
	var namespaces []string
	if len(*targetNamespaces) > 0 {
		namespaces = strings.Split(*targetNamespaces, ",")
	}

	manifest.GenForK8s(g, &manifest.Options{
		Options: service.Options{
			PayloadSize:     int(payloadZ.Value()),
			UploadSize:      int(uploadZ.Value()),
			Timeout:         *clientSideTimeout,
			QueryInParallel: *QueryInParallel,
			LongConn:        *longConn,
		},
		TrafficGenOptions: manifest.TrafficGenOptions{
			NumConcurrentProc: *numTrafficGenProc,
			QueryInterval:     *trafficGenQueryInterval,
		},
		Output:             out,
		Namespaces:         namespaces,
		ReplicaNumberRange: [2]int{1, *maxReplicas},
		Image:              *image,
		CPURequest:         *cpuRequest,
		CPULimit:           *cpuLimit,
	})
}
