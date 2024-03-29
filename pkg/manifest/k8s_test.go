package manifest_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	"github.com/warm-metal/ms-demo-gen.git/pkg/manifest"
	"github.com/warm-metal/ms-demo-gen.git/pkg/service"

	rands "github.com/xyproto/randomstring"
)

func TestMain(m *testing.M) {
	rands.Seed()
	os.Exit(m.Run())
}

func TestManifestGeneration(t *testing.T) {
	dagOpts := &dag.Options{
		NumberVertices:      3,
		InDegreeRange:       [2]int{1, 2},
		OutDegreeRange:      [2]int{0, 2},
		NumberVersionsRange: [2]int{1, 2},
		LongestWalk:         3,
	}

	g := dag.New(dagOpts)
	s := []*strings.Builder{{}, {}}
	manifestOpts := &manifest.Options{
		Options: service.Options{
			PayloadSize:     10,
			UploadSize:      5,
			QueryInParallel: true,
			Address:         ":80",
		},
		Outputs:            make([]io.Writer, len(s)),
		Namespaces:         []string{"ms", "ms2", "ms3"},
		ReplicaNumberRange: [2]int{1, 3},
		Image:              "docker.io/warmmetal/ms-demo-service:latest",
	}

	for i := range s {
		manifestOpts.Outputs[i] = s[i]
	}
	manifest.GenForK8s(g, manifestOpts)
	if len(s[0].String()) == 0 {
		t.Log(s[0].String())
		t.FailNow()
	}
	if len(s[1].String()) == 0 {
		t.Log(s[1].String())
		t.FailNow()
	}

	manifestOpts.CPURequest = "500m"
	manifestOpts.CPULimit = "500m"
	manifest.GenForK8s(g, manifestOpts)
	if len(s[0].String()) == 0 {
		t.Log(s[0].String())
		t.FailNow()
	}
	if len(s[1].String()) == 0 {
		t.Log(s[1].String())
		t.FailNow()
	}

	manifestOpts.CPURequest = "500m"
	manifestOpts.CPULimit = ""
	manifest.GenForK8s(g, manifestOpts)
	if len(s[0].String()) == 0 {
		t.Log(s[0].String())
		t.FailNow()
	}
	if len(s[1].String()) == 0 {
		t.Log(s[1].String())
		t.FailNow()
	}

	manifestOpts.CPULimit = "500m"
	manifestOpts.CPURequest = ""
	manifest.GenForK8s(g, manifestOpts)
	if len(s[0].String()) == 0 {
		t.Log(s[0].String())
		t.FailNow()
	}
	if len(s[1].String()) == 0 {
		t.Log(s[1].String())
		t.FailNow()
	}
}
