package depgraph

import "testing"

func Test_RemoveMedial(t *testing.T) {
	g := Graph{nodes: []*Node{}}
	a := &Node{Name: "A"}
	b := &Node{Name: "B"}
	c := &Node{Name: "C"}

	g.nodes = append(g.nodes, a)
	g.nodes = append(g.nodes, b)
	g.nodes = append(g.nodes, c)

	g.Remove(b)

	for _, n := range g.nodes {
		if n.Name == b.Name {
			t.Fatalf("Found deleted node: %s", n.Name)
		}
	}
}

func Test_RemoveFinal(t *testing.T) {
	g := Graph{nodes: []*Node{}}
	a := &Node{Name: "A"}
	b := &Node{Name: "B"}
	c := &Node{Name: "C"}

	g.nodes = append(g.nodes, a)
	g.nodes = append(g.nodes, b)
	g.nodes = append(g.nodes, c)

	g.Remove(c)

	for _, n := range g.nodes {
		if n.Name == c.Name {
			t.Fatalf("Found deleted node: %s", c.Name)
		}
	}
}
