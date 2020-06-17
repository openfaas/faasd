package pkg

import (
	"log"
)

func buildInstallOrder(svcs []Service) []string {
	graph := Graph{nodes: []*Node{}}

	nodeMap := map[string]*Node{}
	for _, s := range svcs {
		n := &Node{Name: s.Name}
		nodeMap[s.Name] = n
		graph.nodes = append(graph.nodes, n)
	}

	for _, s := range svcs {
		for _, d := range s.DependsOn {
			nodeMap[s.Name].Edges = append(nodeMap[s.Name].Edges, nodeMap[d])
		}
	}

	resolved := &Graph{}
	unresolved := &Graph{}
	for _, g := range graph.nodes {
		resolve(g, resolved, unresolved)
	}

	log.Printf("Start-up order:\n")
	order := []string{}
	for _, node := range resolved.nodes {
		log.Printf("- %s\n", node.Name)
		order = append(order, node.Name)
	}

	return order
}
