package pkg

import (
	"log"

	"github.com/openfaas/faasd/pkg/depgraph"
)

func buildDeploymentOrder(svcs []Service) []string {

	graph := buildServiceGraph(svcs)

	order := graph.Resolve()

	log.Printf("Start-up order:\n")
	for _, node := range order {
		log.Printf("- %s\n", node)
	}

	return order
}

func buildServiceGraph(svcs []Service) *depgraph.Graph {
	graph := depgraph.NewDepgraph()

	nodeMap := map[string]*depgraph.Node{}
	for _, s := range svcs {
		n := &depgraph.Node{Name: s.Name}
		nodeMap[s.Name] = n
		graph.Add(n)

	}

	for _, s := range svcs {
		for _, d := range s.DependsOn {
			nodeMap[s.Name].Edges = append(nodeMap[s.Name].Edges, nodeMap[d])
		}
	}

	return graph
}
