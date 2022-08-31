package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	"github.com/warm-metal/ms-demo-gen.git/pkg/manifest"
	"gonum.org/v1/gonum/graph/encoding/dot"

	rands "github.com/xyproto/randomstring"
)

var (
	numServices        = flag.Int("services", 1, "Number of services in the demo")
	maxCaller          = flag.Int("max-caller", 1, "Maximum number of callers for each service except the root service")
	maxCallee          = flag.Int("max-callee", 1, "Maximum number of callees for each service except leaf services")
	maxReplicas        = flag.Int("max-replicas", 1, "Maximum number of replicas")
	longestCallChain   = flag.Int("longest-call-chain", -1, "Number of services in the longest call chain. -1 means not limit")
	outputDir          = flag.String("out", "", "The directory to where manifests to be generated")
	targetNamespaces   = flag.String("namespaces", "", "Namespaces where workloads to be deployed. Multiple namespaces should be seperated by a comma(,).")
	image              = flag.String("image", "docker.io/warmmetal/ms-demo-service:latest", "Image for each workload")
	alsoOutputTopology = flag.Bool("gen-topology", true, "Output the topology in a DOT file.")
)

func main() {
	rands.Seed()
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

	manifest.GenForK8s(g, &manifest.Options{
		OutputDir:          *outputDir,
		Namespaces:         strings.Split(*targetNamespaces, ","),
		ReplicaNumberRange: [2]int{1, *maxReplicas},
		Image:              *image,
	})
}
