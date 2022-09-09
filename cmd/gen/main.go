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
	numServices             = flag.Int("services", 10, "Number of services to be generated")
	maxCaller               = flag.Int("max-downstream", 2, "Maximum number of downstream services of each service")
	maxCallee               = flag.Int("max-upstream", 3, "Maximum number of upstream services of each service")
	maxReplicas             = flag.Int("max-replicas", 1, "Maximum number of workload replicas for each service")
	maxVersions             = flag.Int("max-versions", 2, "Maximum number of versions of each service")
	longestCallChain        = flag.Int("longest-call-chain", -1, "Number of services in the longest call chain. -1 means no limit")
	outputDir               = flag.String("o", "", "The directory to where manifests to be generated. The default output position is stdout.")
	targetNamespaces        = flag.String("namespaces", "", "Namespaces where workloads to be deployed. Multiple namespaces should be seperated by comma(,). and namespaces should be created manually.")
	image                   = flag.String("image", "docker.io/warmmetal/ms-demo-service:latest", "Image for workloads")
	alsoOutputTopology      = flag.Bool("gen-connectivity", true, "Generate connectivity layout in a DOT file.")
	payloadSize             = flag.String("payload-size", "64", "The payload size of a single query. Such as 10Ki, 1Mi")
	uploadSize              = flag.String("upload-size", "0", "The uploading data size of a single request. Such as 1Ki. If greater than 0, POST requests are issued. Otherwize, GET queries instead.")
	clientSideTimeout       = flag.Duration("timeout", 0, "Client side timeout in time.Duration. 0 means never expired.")
	QueryInParallel         = flag.Bool("parallel", false, "If true, requests to all upstreams are issued concurrently. Otherwise, in the order of upstream setting.")
	longConn                = flag.Bool("long", false, "If true, all queries to the same upstream share the same L4 connection. Otherwise, a particular connection for a single request.")
	numTrafficGenProc       = flag.Int("traffic-gen-proc", 1, "Number of concurrent processors in the traffic generator.")
	trafficGenQueryInterval = flag.Duration("traffic-gen-query-interval", time.Second, "Interval between queries of the traffic generator.")
	cpuRequest              = flag.String("service-cpu-request", "", "CPU requested for each service.")
	cpuLimit                = flag.String("service-cpu-limit", "", "CPU limited for each service.")
	showVersion             = flag.Bool("v", false, "Show the current version")

	Version = ""
)

func main() {
	rands.Seed()
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		return
	}

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

		if err = ioutil.WriteFile(filepath.Join(*outputDir, fmt.Sprintf("connectivity-layout-%s.dot", app)), dotBin, 0755); err != nil {
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
