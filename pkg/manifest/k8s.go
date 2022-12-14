package manifest

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/service"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

type VersionedService struct {
	TrafficGenOptions
	service.Options

	Name        string
	Namespace   string
	App         string
	Version     string
	NumVersions int
	NumReplicas int
	Image       string
	cpuRequest  string
	cpuLimit    string
}

func (s VersionedService) JoinUpstreams() string {
	return strings.Join(s.Upstream, ",")
}

func (s VersionedService) QueryInParallelInInt() int {
	if s.QueryInParallel {
		return 1
	} else {
		return 0
	}
}

func (s VersionedService) LongConnInInt() int {
	if s.LongConn {
		return 1
	} else {
		return 0
	}
}

func (s VersionedService) HasResourceConstraints() bool {
	return len(s.cpuLimit) > 0 || len(s.cpuRequest) > 0
}

func (s VersionedService) CPURequest() string {
	if len(s.cpuRequest) > 0 {
		return s.cpuRequest
	}

	if len(s.cpuLimit) > 0 {
		return "0"
	}

	panic("unreached")
}

func (s VersionedService) CPULimit() string {
	if len(s.cpuLimit) > 0 {
		return s.cpuLimit
	}

	if len(s.cpuRequest) > 0 {
		return s.cpuRequest
	}

	panic("unreached")
}

type TrafficGenOptions struct {
	NumConcurrentProc int
	QueryInterval     time.Duration
}

type Options struct {
	TrafficGenOptions
	service.Options

	Outputs            []io.Writer
	Namespaces         []string
	ReplicaNumberRange [2]int
	Image              string
	CPURequest         string
	CPULimit           string
	App                string

	namespaceMap map[string]string
}

func (o *Options) Namespace(app string) string {
	if len(o.Namespaces) == 0 {
		return "default"
	}

	if len(o.Namespaces) == 1 {
		return o.Namespaces[0]
	}

	if o.namespaceMap == nil {
		o.namespaceMap = make(map[string]string)
	}

	if ns, found := o.namespaceMap[app]; found {
		return ns
	}

	ns := o.Namespaces[rand.Intn(len(o.Namespaces))]
	o.namespaceMap[app] = ns
	return ns
}

func (o Options) NumReplicas() int {
	if o.ReplicaNumberRange[0] <= 0 || o.ReplicaNumberRange[0] > o.ReplicaNumberRange[1] {
		panic(o.ReplicaNumberRange)
	}

	if o.ReplicaNumberRange[0] == o.ReplicaNumberRange[1] {
		return o.ReplicaNumberRange[0]
	}

	return o.ReplicaNumberRange[0] + rand.Intn(o.ReplicaNumberRange[1]-o.ReplicaNumberRange[0])
}

func (o Options) OutputForVersions(numVersions int) []io.Writer {
	if len(o.Outputs) <= numVersions {
		return o.Outputs
	}

	out := make([]io.Writer, len(o.Outputs))
	copy(out, o.Outputs)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out[:numVersions]
}

func parseIntOrDie(v string) (i int) {
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return
}

func (o Options) NewService(node graph.Node) *VersionedService {
	var name, app, version string
	versions := 1
	if node != nil {
		if n, ok := node.(dot.Node); ok {
			name = n.DOTID()
		} else {
			panic("unknown node")
		}

		if n, ok := node.(encoding.Attributer); ok {
			attrs := n.Attributes()
			for _, attr := range attrs {
				switch attr.Key {
				case "app":
					app = attr.Value
				case "version":
					version = attr.Value
				case "versions":
					versions = parseIntOrDie(attr.Value)
				}
			}
		} else {
			panic("unknown node")
		}
	}

	return &VersionedService{
		Options:           o.Options,
		TrafficGenOptions: o.TrafficGenOptions,
		Name:              name,
		App:               app,
		Namespace:         o.Namespace(app),
		Version:           version,
		NumVersions:       versions,
		NumReplicas:       o.NumReplicas(),
		Image:             o.Image,
		cpuRequest:        o.CPURequest,
		cpuLimit:          o.CPULimit,
	}
}

func GenForK8s(g graph.Directed, opts *Options) {
	opts.Address = ":80"
	it := g.Nodes()
	versionMap := make(map[int64]*VersionedService, it.Len())
	serviceMap := make(map[string][]int64)
	for it.Next() {
		from := it.Node()
		fromService := versionMap[from.ID()]
		if fromService == nil {
			fromService = opts.NewService(from)
			versionMap[from.ID()] = fromService
		}

		serviceMap[fromService.App] = append(serviceMap[fromService.App], from.ID())

		targets := g.From(from.ID())
		toServiceMap := make(map[string]struct{}, targets.Len())
		for targets.Next() {
			to := targets.Node()
			toVersion := versionMap[to.ID()]
			if toVersion == nil {
				toVersion = opts.NewService(to)
				versionMap[to.ID()] = toVersion
			}
			toServiceMap[fmt.Sprintf("%s.%s", toVersion.App, toVersion.Namespace)] = struct{}{}
		}

		for upstream := range toServiceMap {
			fromService.Upstream = append(fromService.Upstream, upstream)
		}
	}

	deploymentTmpl := template.Must(template.New("deploy").Parse(deployTemplate))
	serviceTmpl := template.Must(template.New("service").Parse(serviceTemplate))

	const GatewayApp = "gateway"
	gatewayOutput := opts.OutputForVersions(1)[0]
	gatewaySvc := versionMap[serviceMap[GatewayApp][0]]
	if err := deploymentTmpl.Execute(gatewayOutput, gatewaySvc); err != nil {
		panic(err)
	}
	if err := serviceTmpl.Execute(gatewayOutput, gatewaySvc); err != nil {
		panic(err)
	}
	trafficGen := opts.NewService(nil)
	trafficGen.Name = "traffic-generator"
	trafficGen.Namespace = gatewaySvc.Namespace
	trafficGen.App = trafficGen.Name
	trafficGen.Version = "v1"
	trafficGen.NumVersions = 1
	trafficGen.NumReplicas = 1
	trafficGen.Image = "docker.io/warmmetal/ms-demo-traffic:latest"
	trafficGen.Upstream = []string{GatewayApp}
	trafficGen.PayloadSize = -1
	trafficGen.Address = ""
	if err := deploymentTmpl.Execute(gatewayOutput, trafficGen); err != nil {
		panic(err)
	}

	delete(serviceMap, GatewayApp)
	for _, versions := range serviceMap {
		outputs := opts.OutputForVersions(len(versions))

		for _, output := range outputs {
			if err := serviceTmpl.Execute(output, versionMap[versions[0]]); err != nil {
				panic(err)
			}
		}

		for i, versionID := range versions {
			s := versionMap[int64(versionID)]
			if s == nil {
				panic(versionID)
			}
			if err := deploymentTmpl.Execute(outputs[i%len(outputs)], s); err != nil {
				panic(err)
			}
		}
	}
}
