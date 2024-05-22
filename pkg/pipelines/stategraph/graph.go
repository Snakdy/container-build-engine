package stategraph

import (
	"errors"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"slices"
)

// NodeSet a node in this graph is just a string, so a nodeset is a map whose
// keys are the nodes that are present.
type NodeSet map[string]struct{}

// depmap tracks the nodes that have some dependency relationship to
// some other node, represented by the key of the map.
type depmap map[string]NodeSet

type Graph struct {
	nodes NodeSet

	// Maintain dependency relationships in both directions. These
	// data structures are the edges of the graph.

	// `dependencies` tracks child -> parents.
	dependencies depmap
	// `dependents` tracks parent -> children.
	dependents depmap
	// Keep track of the nodes of the graph themselves.
}

func New() *Graph {
	return &Graph{
		dependencies: make(depmap),
		dependents:   make(depmap),
		nodes:        make(NodeSet),
	}
}

func (g *Graph) DependOn(statement pipelines.OrderedPipelineStatement) error {
	if slices.Contains(statement.DependsOn, statement.ID) {
		return errors.New("self-referential dependencies not allowed")
	}

	for _, d := range statement.DependsOn {
		if g.DependsOn(d, statement.ID) {
			return errors.New("circular dependencies not allowed")
		}
	}

	// add nodes
	for _, parent := range statement.DependsOn {
		g.nodes[parent] = struct{}{}
	}
	g.nodes[statement.ID] = struct{}{}

	// add edges
	for _, dependency := range statement.DependsOn {
		addNodeToNodeset(g.dependents, dependency, statement.ID)
		addNodeToNodeset(g.dependencies, statement.ID, dependency)
	}

	return nil
}

func (g *Graph) DependsOn(child, parent string) bool {
	deps := g.Dependencies(child)
	_, ok := deps[parent]
	return ok
}

func (g *Graph) HasDependent(parent, child string) bool {
	deps := g.Dependents(parent)
	_, ok := deps[child]
	return ok
}

func (g *Graph) Leaves() []string {
	leaves := make([]string, 0)

	for node := range g.nodes {
		if _, ok := g.dependencies[node]; !ok {
			leaves = append(leaves, node)
		}
	}

	return leaves
}

// TopoSortedLayers returns a slice of all the graph nodes in topological sort order. That is,
// if `B` depends on `A`, then `A` is guaranteed to come before `B` in the sorted output.
// The graph is guaranteed to be cycle-free because cycles are detected while building the
// graph. Additionally, the output is grouped into "layers", which are guaranteed to not have
// any dependencies within each layer. This is useful, e.g. when building an execution plan for
// some DAG, in which case each element within each layer could be executed in parallel. If you
// do not need this layered property, use `Graph.TopoSorted()`, which flattens all elements.
func (g *Graph) TopoSortedLayers() [][]string {
	var layers [][]string

	// Copy the graph
	shrinkingGraph := g.clone()
	for {
		leaves := shrinkingGraph.Leaves()
		if len(leaves) == 0 {
			break
		}

		layers = append(layers, leaves)
		for _, leafNode := range leaves {
			shrinkingGraph.remove(leafNode)
		}
	}

	return layers
}

func removeFromDepmap(dm depmap, key, node string) {
	nodes := dm[key]
	if len(nodes) == 1 {
		// The only element in the nodeset must be `node`, so we
		// can delete the entry entirely.
		delete(dm, key)
	} else {
		// Otherwise, remove the single node from the nodeset.
		delete(nodes, node)
	}
}

func (g *Graph) remove(node string) {
	// Remove edges from things that depend on `node`.
	for dependent := range g.dependents[node] {
		removeFromDepmap(g.dependencies, dependent, node)
	}
	delete(g.dependents, node)

	// Remove all edges from node to the things it depends on.
	for dependency := range g.dependencies[node] {
		removeFromDepmap(g.dependents, dependency, node)
	}
	delete(g.dependencies, node)

	// Finally, remove the node itself.
	delete(g.nodes, node)
}

// TopoSorted returns all the nodes in the graph in topological sort order.
// See also `Graph.TopoSortedLayers()`.
func (g *Graph) TopoSorted() []string {
	nodeCount := 0
	layers := g.TopoSortedLayers()
	for _, layer := range layers {
		nodeCount += len(layer)
	}

	allNodes := make([]string, 0, nodeCount)
	for _, layer := range layers {
		for _, node := range layer {
			allNodes = append(allNodes, node)
		}
	}

	return allNodes
}

func (g *Graph) Dependencies(child string) NodeSet {
	return g.buildTransitive(child, g.immediateDependencies)
}

func (g *Graph) immediateDependencies(node string) NodeSet {
	return g.dependencies[node]
}

func (g *Graph) Dependents(parent string) NodeSet {
	return g.buildTransitive(parent, g.immediateDependents)
}

func (g *Graph) immediateDependents(node string) NodeSet {
	return g.dependents[node]
}

func (g *Graph) clone() *Graph {
	return &Graph{
		dependencies: copyDepmap(g.dependencies),
		dependents:   copyDepmap(g.dependents),
		nodes:        copyNodeset(g.nodes),
	}
}

// buildTransitive starts at `root` and continues calling `nextFn` to keep discovering more nodes until
// the graph cannot produce anymore. It returns the set of all discovered nodes.
func (g *Graph) buildTransitive(root string, nextFn func(string) NodeSet) NodeSet {
	if _, ok := g.nodes[root]; !ok {
		return nil
	}

	out := make(NodeSet)
	searchNext := []string{root}
	for len(searchNext) > 0 {
		// List of new nodes from this layer of the dependency graph. This is
		// assigned to `searchNext` at the end of the outer "discovery" loop.
		var discovered []string
		for _, node := range searchNext {
			// For each node to discover, find the next nodes.
			for nextNode := range nextFn(node) {
				// If we have not seen the node before, add it to the output as well
				// as the list of nodes to traverse in the next iteration.
				if _, ok := out[nextNode]; !ok {
					out[nextNode] = struct{}{}
					discovered = append(discovered, nextNode)
				}
			}
		}
		searchNext = discovered
	}

	return out
}

func copyNodeset(s NodeSet) NodeSet {
	out := make(NodeSet, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func copyDepmap(m depmap) depmap {
	out := make(depmap, len(m))
	for k, v := range m {
		out[k] = copyNodeset(v)
	}
	return out
}

func addNodeToNodeset(dm depmap, key, node string) {
	nodes, ok := dm[key]
	if !ok {
		nodes = make(NodeSet)
		dm[key] = nodes
	}
	nodes[node] = struct{}{}
}
