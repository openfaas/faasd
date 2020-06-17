package pkg

import "log"

// Node represents a node in a Graph with
// 0 to many edges
type Node struct {
	Name  string
	Edges []*Node
}

// Graph is a collection of nodes
type Graph struct {
	nodes []*Node
}

// Contains returns true if the target Node is found
// in its list
func (g *Graph) Contains(target *Node) bool {
	for _, g := range g.nodes {
		if g.Name == target.Name {
			return true
		}
	}

	return false
}

// Add places a Node into the current Graph
func (g *Graph) Add(target *Node) {
	g.nodes = append(g.nodes, target)
}

// Remove deletes a target Node reference from the
// list of nodes in the graph
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

// resolve finds the order of dependencies for a graph
// of nodes.
// Inspired by algorithm from
// https://www.electricmonk.nl/log/2008/08/07/dependency-resolving-algorithm/
func resolve(node *Node, resolved, unresolved *Graph) {
	unresolved.Add(node)

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

	resolved.Add(node)
	unresolved.Remove(node)
}
