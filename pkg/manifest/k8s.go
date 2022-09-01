package manifest

import (
	"io"
	"math/rand"
	"strings"
	"text/template"

	"github.com/warm-metal/ms-demo-gen.git/pkg/service"
	"gonum.org/v1/gonum/graph"

	rands "github.com/xyproto/randomstring"
)

type Service struct {
	service.Options

	Name        string
	Namespace   string
	NumReplicas int
	Image       string
}

func (s Service) JoinUpstreams() string {
	return strings.Join(s.Upstream, ",")
}

type Options struct {
	service.Options

	Output             io.Writer
	Namespaces         []string
	ReplicaNumberRange [2]int
	Image              string
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

func (o Options) NewService() *Service {
	return &Service{
		Options:     o.Options,
		Name:        rands.HumanFriendlyEnglishString(10),
		Namespace:   o.Namespace(),
		NumReplicas: o.NumReplicas(),
		Image:       o.Image,
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
			fromService = opts.NewService()
			serviceMap[from.ID()] = fromService
		}

		targets := g.From(from.ID())
		for targets.Next() {
			to := targets.Node()
			toService := serviceMap[to.ID()]
			if toService == nil {
				toService = opts.NewService()
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
	trafficGen := opts.NewService()
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
