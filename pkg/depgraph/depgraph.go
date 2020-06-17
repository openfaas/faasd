package depgraph

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

func NewDepgraph() *Graph {
	return &Graph{
		nodes: []*Node{},
	}
}

// Nodes returns the nodes within the graph
func (g *Graph) Nodes() []*Node {
	return g.nodes
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

// Resolve retruns a list of node names in order of their dependencies.
// A use case may be for determining the correct order to install
// software packages, or to start services.
// Based upon the algorithm described by Ferry Boender in the following article
// https://www.electricmonk.nl/log/2008/08/07/dependency-resolving-algorithm/
func (g *Graph) Resolve() []string {
	resolved := &Graph{}
	unresolved := &Graph{}
	for _, node := range g.nodes {
		resolve(node, resolved, unresolved)
	}

	order := []string{}

	for _, node := range resolved.Nodes() {
		order = append(order, node.Name)
	}

	return order
}

// resolve mutates the resolved graph for a given starting
// node. The unresolved graph is used to detect a circular graph
// error and will throw a panic. This can be caught with a resolve
// in a go routine.
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
