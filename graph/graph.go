package graph

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

const dotTemplate = `strict {{.GraphType}} {
{{range $k, $v := .Attributes}}
	{{$k}}="{{$v}}";
{{end}}
{{range $s := .Statements}}
	"{{.Source}}" {{if .Target}}{{$.EdgeOperator}} "{{.Target}}" [ label="{{.EdgeLabel}}", weight={{.EdgeWeight}} ]{{else}}[ label="{{.VertexLabel}}", weight={{.SourceWeight}} ]{{end}};
{{end}}
}
`

type Metadata string

type Graph[EdgePropertyType any, VertexPropertyType any] interface {
	AddEdge(edge Edge[EdgePropertyType, VertexPropertyType])
	AddVertex(vertex Vertex[VertexPropertyType])
	DeleteEdge(edge Edge[EdgePropertyType, VertexPropertyType])
	DeleteVertex(vertex Vertex[VertexPropertyType])
	GetOutEdges(vertex Vertex[VertexPropertyType]) []Edge[EdgePropertyType, VertexPropertyType]
	GetInEdges(vertex Vertex[VertexPropertyType]) []Edge[EdgePropertyType, VertexPropertyType]
	GetAllVertices() []Vertex[VertexPropertyType]
	GetAllEdges() []Edge[EdgePropertyType, VertexPropertyType]
	SetMetadata(key Metadata, val any)
	GetMetadata(key Metadata) any
	GetAllMetadata() map[Metadata]any
	GetVertexById(id string) Vertex[VertexPropertyType]
	GetEdgeById(id string) Edge[EdgePropertyType, VertexPropertyType]
}

type Edge[EdgePropertyType any, VertexPropertyType any] interface {
	SetID(id string)
	SetFrom(vertex Vertex[VertexPropertyType])
	SetTo(vertex Vertex[VertexPropertyType])
	SetProperty(key string, val EdgePropertyType)
	GetID() string
	GetFrom() Vertex[VertexPropertyType]
	GetTo() Vertex[VertexPropertyType]
	GetProperty(key string) EdgePropertyType
	GetAllProperties() map[string]EdgePropertyType
}

type Vertex[VertexPropertyType any] interface {
	SetID(id string)
	SetProperty(key string, val VertexPropertyType)
	GetID() string
	GetProperty(key string) VertexPropertyType
	GetAllProperties() map[string]VertexPropertyType
}

// Clone Ept: EdgePropertyType, Vpt: VertexPropertyType
func Clone[Ept any, Vpt any](graph Graph[Ept, Vpt], newGraph func() Graph[Ept, Vpt], newEdge func() Edge[Ept, Vpt], newVertex func() Vertex[Vpt]) Graph[Ept, Vpt] {
	// Use the provided factory function to create a new graph instance
	clonedGraph := newGraph()
	for k, v := range graph.GetAllMetadata() {
		clonedGraph.SetMetadata(k, v)
	}

	// Create a map to track the mapping from original vertices to cloned vertices
	vertexMap := make(map[string]Vertex[Vpt])

	// Clone all vertices
	for _, v := range graph.GetAllVertices() {
		clonedVertex := newVertex() // Use the factory function to create a new vertex instance
		clonedVertex.SetID(v.GetID())
		// Retrieve and set all properties
		for key, val := range v.GetAllProperties() {
			clonedVertex.SetProperty(key, val)
		}
		// Add to the new graph and update the map
		clonedGraph.AddVertex(clonedVertex)
		vertexMap[v.GetID()] = clonedVertex
	}

	// Clone all edges
	for _, e := range graph.GetAllEdges() {
		clonedEdge := newEdge() // Use the factory function to create a new edge instance
		clonedEdge.SetID(e.GetID())
		// Set the start and end points, using the map to find the corresponding cloned vertices
		clonedEdge.SetFrom(vertexMap[e.GetFrom().GetID()])
		clonedEdge.SetTo(vertexMap[e.GetTo().GetID()])
		// Retrieve and set all properties
		for key, val := range e.GetAllProperties() {
			clonedEdge.SetProperty(key, val)
		}
		if clonedEdge.GetFrom() == nil || clonedEdge.GetTo() == nil {
			fmt.Println("nil")
		}
		// Add to the new graph
		clonedGraph.AddEdge(clonedEdge)
	}

	// Return the cloned graph
	return clonedGraph
}

func Visualize[EdgePropertyType any, VertexPropertyType any](graph Graph[EdgePropertyType, VertexPropertyType], filename string, f func(vertex Vertex[VertexPropertyType]) string) error {
	desc, err := generateDOT(graph, f)
	if err != nil {
		return fmt.Errorf("failed to generate DOT description: %w", err)
	}
	w, _ := os.Create(filename)
	return renderDOT(w, desc)
}

type description[PropertyType any] struct {
	GraphType    string
	Attributes   map[string]string
	EdgeOperator string
	Statements   []statement[PropertyType]
}

type statement[PropertyType any] struct {
	Source       interface{}
	Target       interface{}
	SourceWeight int
	EdgeLabel    string
	EdgeWeight   int
	VertexLabel  string
}

// design flaw: only vertex property can be shown
func generateDOT[EdgePropertyType any, VertexPropertyType any](g Graph[EdgePropertyType, VertexPropertyType], f func(node Vertex[VertexPropertyType]) string) (description[VertexPropertyType], error) {
	desc := description[VertexPropertyType]{
		GraphType:    "graph",
		Attributes:   make(map[string]string),
		EdgeOperator: "--",
		Statements:   make([]statement[VertexPropertyType], 0),
	}
	if f == nil {
		f = func(node Vertex[VertexPropertyType]) string {
			return node.GetID()
		}
	}

	desc.GraphType = "digraph"
	desc.EdgeOperator = "->"

	for _, vertex := range g.GetAllVertices() {
		stmt := statement[VertexPropertyType]{
			Source:       vertex.GetID(),
			SourceWeight: 1,
			VertexLabel:  f(vertex),
		}
		desc.Statements = append(desc.Statements, stmt)

		for _, edge := range g.GetOutEdges(vertex) {
			stmt1 := statement[VertexPropertyType]{
				Source:     vertex.GetID(),
				Target:     edge.GetTo().GetID(),
				EdgeWeight: 1,
				//EdgeLabel:  f(edge),
			}
			desc.Statements = append(desc.Statements, stmt1)
		}
	}

	return desc, nil
}

func renderDOT[PropertyType any](w io.Writer, d description[PropertyType]) error {
	tpl, err := template.New("dotTemplate").Parse(dotTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tpl.Execute(w, d)
}
