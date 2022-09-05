package manifest_test

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	"github.com/warm-metal/ms-demo-gen.git/pkg/manifest"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"

	rands "github.com/xyproto/randomstring"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	rands.Seed()
	os.Exit(m.Run())
}

func TestManifestGeneration(t *testing.T) {
	dagOpts := &dag.Options{
		NumberVertices: 3,
		InDegreeRange:  [2]int{1, 2},
		OutDegreeRange: [2]int{0, 2},
		LongestWalk:    3,
	}

	g := dag.New(dagOpts)
	s := &strings.Builder{}
	manifestOpts := &manifest.Options{
		Options: service.Options{
			PayloadSize:     10,
			UploadSize:      5,
			QueryInParallel: true,
			Address:         ":80",
		},
		Output:             s,
		Namespaces:         []string{"ms", "ms2", "ms3"},
		ReplicaNumberRange: [2]int{1, 3},
		Image:              "docker.io/warmmetal/ms-demo-service:latest",
		CPURequest:         "1000m",
		CPULimit:           "1000m",
	}
	manifest.GenForK8s(g, manifestOpts)
	t.Log(s.String())
}
