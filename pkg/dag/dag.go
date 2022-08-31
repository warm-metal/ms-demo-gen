package dag

import (
	"math/rand"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/simple"
)

type Options struct {
	NumberVertices int
	InDegreeRange  [2]int
	OutDegreeRange [2]int
	LongestWalk    int
}

type Service struct {
	id        int64
	outDegree int
	depth     int
}

func (s *Service) ID() int64 {
	return s.id
}

func (s *Service) IsRoot() bool {
	return s.id == 1
}

func (s *Service) DOTAttributers() (graph, node, edge encoding.Attributer) {
	return nil, &encoding.Attributes{
		{
			Key:   "svc",
			Value: "a",
		},
		{
			Key:   "svc",
			Value: "b",
		},
	}, nil
}

func calcVerticesHaveOutDegrees(vertices []*Service, upperBoundOutDegree, maxDepth int) []*Service {
	maxDepth -= 1
	available := make([]*Service, 0, len(vertices))
	for _, v := range vertices {
		if v.outDegree > upperBoundOutDegree {
			panic(vertices)
		}

		if maxDepth > 0 && v.depth >= maxDepth {
			continue
		}

		if v.outDegree < upperBoundOutDegree {
			available = append(available, v)
		}
	}

	return available
}

func selectVerticesRandomly(vertices []*Service, numTargts int) []*Service {
	if numTargts == 0 {
		panic("invalid in degree 0")
	}

	targets := make([]*Service, numTargts)
	for i := 0; i < numTargts; i++ {
		index := rand.Intn(len(vertices))
		targets[i] = vertices[index]
		vertices = append(vertices[:index], vertices[index+1:]...)
	}

	return targets
}

func New(opts *Options) graph.Directed {
	g := simple.NewDirectedGraph()
	vertices := make([]*Service, opts.NumberVertices)
	for i := 1; i <= opts.NumberVertices; i++ {
		vertex := &Service{id: int64(i)}
		vertices[i-1] = vertex
		if vertex.IsRoot() {
			g.AddNode(vertex)
			continue
		}

		availableVertices := calcVerticesHaveOutDegrees(vertices[:i-1], opts.OutDegreeRange[1], opts.LongestWalk)
		// FIXME remove redundant walk paths.
		if len(availableVertices) == 0 {
			panic("all vertices have no more out degrees")
		}

		upperBoundInDegree := opts.InDegreeRange[1]
		if len(availableVertices) < upperBoundInDegree {
			upperBoundInDegree = len(availableVertices)
		}

		if upperBoundInDegree < opts.InDegreeRange[0] {
			panic("upperBoundInDegree is lower than the lower bound")
		}

		inDegree := 0
		if upperBoundInDegree == opts.InDegreeRange[0] {
			inDegree = upperBoundInDegree
		} else {
			inDegree = rand.Intn(upperBoundInDegree-opts.InDegreeRange[0]) + opts.InDegreeRange[0]
		}

		fromVertices := selectVerticesRandomly(availableVertices, inDegree)
		if len(fromVertices) == 0 {
			panic(`no source vertex found for target vertex`)
		}
		for _, v := range fromVertices {
			v.outDegree += 1
			depth := v.depth + 1
			if depth > vertex.depth {
				vertex.depth = depth
			}

			g.SetEdge(g.NewEdge(v, vertex))
		}
	}

	return g
}
