package dag

import (
	"fmt"
	"math/rand"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/simple"
)

type Options struct {
	NumberVertices      int
	InDegreeRange       [2]int
	OutDegreeRange      [2]int
	NumberVersionsRange [2]int
	LongestWalk         int
}

type Service struct {
	id        int64
	outDegree int
	depth     int
	versions  int
	app       string
	version   string
}

func (s *Service) ID() int64 {
	return s.id
}

func (s *Service) DOTID() string {
	return fmt.Sprintf("%s-%s", s.app, s.version)
}

func (s *Service) IsRoot() bool {
	return s.id == 1
}

func (s *Service) Attributes() []encoding.Attribute {
	return []encoding.Attribute{
		{Key: "app", Value: s.app},
		{Key: "version", Value: s.version},
		{Key: "versions", Value: fmt.Sprintf("%d", s.versions)},
	}
}

func createServices(id int64, name string, numVersions int) (nextID int64, svcs []Service) {
	svcs = make([]Service, numVersions)
	nextID = id
	for i := 0; i < numVersions; i++ {
		svcs[i] = Service{
			id:       nextID,
			app:      name,
			version:  fmt.Sprintf("v%d", i+1),
			versions: numVersions,
		}
		nextID++
	}
	return
}

type ServiceGraph struct {
	*simple.DirectedGraph
}

func (s *ServiceGraph) DOTAttributers() (graph, node, edge encoding.Attributer) {
	return &encoding.Attributes{{Key: "rankdir", Value: `"LR"`}},
		&encoding.Attributes{{Key: "shape", Value: "box"}}, nil
}

func calcVerticesHaveOutDegrees(vertices []Service, upperBoundOutDegree, maxDepth int) []*Service {
	maxDepth -= 1
	available := make([]*Service, 0, len(vertices))
	for i := range vertices {
		v := &vertices[i]
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
	g := &ServiceGraph{
		DirectedGraph: simple.NewDirectedGraph(),
	}

	vertices := make([]Service, 0, opts.NumberVertices*opts.NumberVersionsRange[1])
	nextVertexID := int64(1)
	gen := createNameGen()

	for i := 1; i <= opts.NumberVertices; i++ {
		var fromVertices []*Service
		if len(vertices) > 0 {
			availableVertices := calcVerticesHaveOutDegrees(vertices, opts.OutDegreeRange[1], opts.LongestWalk)
			if len(availableVertices) == 0 {
				panic(fmt.Sprint(vertices))
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
				// Plusing 1 to include the upper bound
				inDegree = rand.Intn(upperBoundInDegree-opts.InDegreeRange[0]+1) + opts.InDegreeRange[0]
			}

			fromVertices = selectVerticesRandomly(availableVertices, inDegree)
			if len(fromVertices) == 0 {
				panic(`no source vertex found for target vertex`)
			}
		}

		versions := 1
		if len(vertices) > 0 {
			if opts.NumberVersionsRange[0] == opts.NumberVersionsRange[1] {
				versions = opts.NumberVersionsRange[0]
			} else {
				versions = rand.Intn(opts.NumberVersionsRange[1]-opts.NumberVersionsRange[0]+1) + opts.NumberVersionsRange[0]
			}
		}

		var newVertices []Service
		var name string
		if nextVertexID == 1 {
			name = "gateway"
		} else {
			name = gen.Name()
		}

		nextVertexID, newVertices = createServices(nextVertexID, name, versions)
		newVertexIndex := len(vertices)
		vertices = append(vertices, newVertices...)

		for _, v := range fromVertices {
			v.outDegree += 1
			for ; newVertexIndex < len(vertices); newVertexIndex++ {
				vertex := &vertices[newVertexIndex]
				depth := v.depth + 1
				if depth > vertex.depth {
					vertex.depth = depth
				}

				g.SetEdge(g.NewEdge(v, vertex))
			}
		}
	}

	return g
}
