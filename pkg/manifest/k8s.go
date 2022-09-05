package manifest

import (
	"io"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/service"
	"gonum.org/v1/gonum/graph"

	rands "github.com/xyproto/randomstring"
)

type Service struct {
	TrafficGenOptions
	service.Options

	Name        string
	Namespace   string
	App         string
	NumReplicas int
	Image       string
	cpuRequest  string
	cpuLimit    string
}

func (s Service) JoinUpstreams() string {
	return strings.Join(s.Upstream, ",")
}

func (s Service) QueryInParallelInInt() int {
	if s.QueryInParallel {
		return 1
	} else {
		return 0
	}
}

func (s Service) LongConnInInt() int {
	if s.LongConn {
		return 1
	} else {
		return 0
	}
}

func (s Service) HasResourceConstraints() bool {
	return len(s.cpuLimit) > 0 || len(s.cpuRequest) > 0
}

func (s Service) CPURequest() string {
	if len(s.cpuRequest) > 0 {
		return s.cpuRequest
	}

	if len(s.cpuLimit) > 0 {
		return "0"
	}

	panic("unreached")
}

func (s Service) CPULimit() string {
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

	Output             io.Writer
	Namespaces         []string
	ReplicaNumberRange [2]int
	Image              string
	CPURequest         string
	CPULimit           string
	App                string
}

func (o Options) Namespace() string {
	if len(o.Namespaces) == 0 {
		return "default"
	}

	if len(o.Namespaces) == 1 {
		return o.Namespaces[0]
	}

	return o.Namespaces[rand.Intn(len(o.Namespaces))]
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

func (o Options) NewService(id int64) *Service {
	name := "gateway"
	if id > 1 {
		name = rands.HumanFriendlyEnglishString(10)
	}

	return &Service{
		Options:           o.Options,
		TrafficGenOptions: o.TrafficGenOptions,
		Name:              name,
		Namespace:         o.Namespace(),
		App:               o.App,
		NumReplicas:       o.NumReplicas(),
		Image:             o.Image,
		cpuRequest:        o.CPURequest,
		cpuLimit:          o.CPULimit,
	}
}

func GenForK8s(g graph.Directed, opts *Options) {
	opts.Address = ":80"
	it := g.Nodes()
	serviceMap := make(map[int64]*Service, it.Len())
	for it.Next() {
		from := it.Node()
		fromService := serviceMap[from.ID()]
		if fromService == nil {
			fromService = opts.NewService(from.ID())
			serviceMap[from.ID()] = fromService
		}

		targets := g.From(from.ID())
		for targets.Next() {
			to := targets.Node()
			toService := serviceMap[to.ID()]
			if toService == nil {
				toService = opts.NewService(to.ID())
				serviceMap[to.ID()] = toService
			}
			fromService.Upstream = append(fromService.Upstream, toService.Name)
		}
	}

	workloadTmpl := template.Must(template.New("workload").Parse(deployTemplate + serviceTemplate))
	for i := 1; i <= len(serviceMap); i++ {
		s := serviceMap[int64(i)]
		if s == nil {
			panic(i)
		}
		if err := workloadTmpl.Execute(opts.Output, s); err != nil {
			panic(err)
		}
	}

	deploymentTmpl := template.Must(template.New("deploy").Parse(deployTemplate))
	trafficGen := opts.NewService(int64(len(serviceMap) + 1))
	trafficGen.Name = "traffic-generator"
	trafficGen.NumReplicas = 1
	trafficGen.Image = "docker.io/warmmetal/ms-demo-traffic:latest"
	trafficGen.Upstream = []string{serviceMap[1].Name}
	trafficGen.PayloadSize = -1
	trafficGen.Address = ""
	if err := deploymentTmpl.Execute(opts.Output, trafficGen); err != nil {
		panic(err)
	}
}
