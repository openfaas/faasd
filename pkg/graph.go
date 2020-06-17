package pkg

import "log"

// resolve finds the order of dependencies for a graph
// of nodes.
// Inspired by algorithm from
// https://www.electricmonk.nl/log/2008/08/07/dependency-resolving-algorithm/
func resolve(node *Node, resolved, unresolved *Graph) {
	unresolved.nodes = append(unresolved.nodes, node)

	for _, edge := range node.Edges {

		if !resolved.Contains(edge) && unresolved.Contains(edge) {
			log.Panicf("edge: %s may be a circular dependency", edge.Name)
		}

		resolve(edge, resolved, unresolved)
	}

	for _, r := range resolved.nodes {
		if r.Name == node.Name {
			return
		}
	}

	resolved.nodes = append(resolved.nodes, node)

	unresolved.Remove(node)
}

func newNode(name string, edges []*Node) *Node {
	return &Node{
		Name:  name,
		Edges: edges,
	}
}

type Node struct {
	Name  string
	Edges []*Node
}

type Graph struct {
	nodes []*Node
}

func (g *Graph) Contains(target *Node) bool {
	for _, g := range g.nodes {
		if g.Name == target.Name {
			return true
		}
	}

	return false
}
func (g *Graph) Remove(target *Node) {
	var found *int
	for i, n := range g.nodes {
		if n == target {
			found = &i
			break
		}
	}

	if found != nil {
		g.nodes = append(g.nodes[:*found], g.nodes[*found+1:]...)
	}
}
