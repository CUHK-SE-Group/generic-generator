package schemas

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/CUHK-SE-Group/generic-generator/graph"
	A "github.com/IBM/fp-go/array"
)

type GrammarType int

// 带yes标记的symbol要指定生成策略
const (
	GrammarProduction GrammarType = 1 << iota
	GrammarOR                     // yes
	GrammarCatenate
	GrammarOptional // yes
	GrammarREP      // yes
	GrammarPLUS     // yes
	GrammarEXT      // yes
	GrammarSUB      // yes
	GrammarID
	GrammarTerminal
	GrammarChoice
)
const (
	Prop     = "Property"
	StartSym = "startSym"
)

var typeStrRep = map[GrammarType]string{
	GrammarProduction: "GrammarProduction",
	GrammarOR:         "GrammarOR",
	GrammarCatenate:   "GrammarCatenate",
	GrammarOptional:   "GrammarOptional",
	GrammarREP:        "GrammarREP",
	GrammarPLUS:       "GrammarPLUS",
	GrammarEXT:        "GrammarEXT",
	GrammarSUB:        "GrammarSUB",
	GrammarID:         "GrammarID",
	GrammarTerminal:   "GrammarTerminal",
	GrammarChoice:     "GrammarChoice",
}

func GetGrammarTypeStr(t GrammarType) string {
	return typeStrRep[t]
}

type Property struct {
	Type               GrammarType
	Gram               *Grammar
	Content            string
	DistanceToTerminal int
}

type Options struct {
	StartSym     string
	LoadFromFile string
}

type Option func(*Options)

func WithStartSym(startSym string) Option {
	return func(o *Options) {
		o.StartSym = startSym
	}
}
func WithLoadFromFile(filename string) Option {
	return func(o *Options) {
		o.LoadFromFile = filename
	}
}

type Grammar struct {
	internal graph.Graph[string, Property]
}

func (g *Grammar) Save(filename string) error {
	data, err := marshalGrammar(g)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func NewGrammar(opts ...Option) *Grammar {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}
	newG := &Grammar{}
	newG.internal = graph.NewGraph[string, Property](graph.WithPersistent(true))
	if options.LoadFromFile != "" {
		file, err := os.ReadFile(options.LoadFromFile)
		if err != nil {
			panic(err)
		}
		newG, err = unmarshalGrammar(file)
	}
	if options.StartSym != "" {
		newG.internal.SetMetadata(StartSym, options.StartSym)
	}
	return newG
}
func (g *Grammar) GetInternal() graph.Graph[string, Property] {
	return g.internal
}
func (g *Grammar) GetStartSym() string {
	return g.internal.GetMetadata(StartSym).(string)
}

func (g *Grammar) GetNode(id string) *Node {
	if inter := g.internal.GetVertexById(id); inter != nil {
		return &Node{internal: inter}
	}
	return nil
}

func (g *Grammar) GetEdge(id string) (*Node, *Node) {
	from, to := ExtractEdgeID(id)
	return g.GetNode(from), g.GetNode(to)
}

func (p *Grammar) MergeProduction() {
	start := p.internal.GetMetadata(StartSym).(string)
	queue := []*Node{p.GetNode(start)}
	visited := make(map[string]any)
	productions := []*Node{p.GetNode(start)}
	for len(queue) != 0 {
		for _, n := range queue[0].GetSymbols() {
			if n.GetType() == GrammarID {
				productions = append(productions, n)
				v := p.GetNode(fmt.Sprintf("%s", n.GetContent()))
				if v != nil {
					n.AddSymbol(v)
					queue = append(queue, v)
				}
			}
			if _, ok := visited[n.GetID()]; !ok {
				queue = append(queue, n)
				visited[n.GetID()] = ""
			}
		}
		queue = queue[1:]
	}
}

func (g *Grammar) BuildShortestNotation() {
	vertices := g.internal.GetAllVertices()
	numVertices := len(vertices)
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].GetID() > vertices[j].GetID()
	})
	vertexMap := make(map[string]int)
	for i, vertex := range vertices {
		vertexMap[vertex.GetID()] = i
	}

	// some sufficiently large value
	// indicating not-reachable
	inf := int(1e8)
	distance := make([]int, numVertices)
	for i, vertex := range vertices {
		if vertex.GetProperty(Prop).Type == GrammarTerminal {
			distance[i] = 0
		} else {
			distance[i] = inf
		}
	}

	// Bellman-ford-like process;
	// this should terminate within O(numVertices) iterations,
	// i.e. no more relaxations after that
	round := 0
	for {
		round++
		stop := true
		for index, current := range vertices {
			adjacent := g.internal.GetOutEdges(current)
			pre := distance[index]
			if current.GetProperty(Prop).Type == GrammarTerminal {
				// do nothing
			} else if current.GetProperty(Prop).Type == GrammarOR {
				//if strings.Contains(current.GetProperty(Prop).Content, "1") {
				//	fmt.Println()
				//}
				// 1 + min of {distances of outgoing neighbors}
				best := inf
				for _, e := range adjacent {
					next_index := vertexMap[e.GetTo().GetID()]
					best = min(best, distance[next_index])
				}
				best += 1
				distance[index] = min(distance[index], best)
			} else {
				// 1 + sum of {distances of outgoing neighbors}
				sum := 0
				for _, e := range adjacent {
					next_index := vertexMap[e.GetTo().GetID()]
					sum += distance[next_index]
					// watch out for overflows
					if distance[next_index] >= inf {
						break
					}
				}
				sum += 1
				distance[index] = min(distance[index], sum)
				if distance[index] < 1 {
					fmt.Println("fuck")
				}
			}
			if pre != distance[index] {
				stop = false
			}
		}
		if stop {
			fmt.Println(round)
			break
		}
	}

	for index, v := range vertices {
		if v.GetProperty(Prop).Type == GrammarTerminal {
			continue
		}
		prop := v.GetProperty(Prop)
		prop.DistanceToTerminal = distance[index]
		v.SetProperty(Prop, prop)
	}
}

type Node struct {
	internal graph.Vertex[Property]
}

func dfs(node *Node, visit func(*Node)) {
	if node == nil {
		return
	}

	// 访问当前节点
	visit(node)

	// 递归遍历子节点
	nodes := node.GetSymbols()
	for _, child := range nodes {
		if child.GetID() == node.GetID() {
			continue
		}
		dfs(child, visit)
	}
}
func (g *Grammar) PrintTerminals(startSym string) {
	root := g.GetNode(startSym)
	if root == nil {
		return
	}

	dfs(root, func(node *Node) {
		if node.GetType() == GrammarTerminal {
			fmt.Printf("%s", strings.Trim(node.GetContent(), "'\""))
		}
	})
	fmt.Println()
}
func NewNode(g *Grammar, tp GrammarType, id, content string) *Node {
	n := graph.NewVertex[Property]()
	n.SetProperty(Prop, Property{
		Type:    tp,
		Gram:    g,
		Content: content,
	})
	n.SetID(id)
	return &Node{internal: n}
}
func (g *Node) newEdge(id string, from, to *Node) graph.Edge[string, Property] {
	res := graph.NewEdge[string, Property]()
	res.SetID(id)
	res.SetFrom(from.internal)
	res.SetTo(to.internal)
	res.SetMeta(g.GetMeta())

	g.SetMeta(g.GetMeta() + 1) // todo: fix this. To align with the snapshot of original node
	return res
}

func (g *Node) Clone(belongto *Grammar) *Node {
	newInternal := graph.CloneVertex(g.internal, graph.NewVertex[Property])
	if belongto != nil {
		p := newInternal.GetProperty(Prop)
		p.Gram = belongto
		newInternal.SetProperty(Prop, p)
	}
	return &Node{internal: newInternal}
}

func (g *Node) GetType() GrammarType {
	if g.internal == nil {
		return 0
	}
	return g.internal.GetProperty(Prop).Type
}

func (g *Node) GetID() string {
	return g.internal.GetID()
}
func (g *Node) SetID(id string) {
	g.internal.SetID(id)
}
func (g *Node) GetMeta() int {
	if g.internal.GetMeta() == nil {
		return 0
	}
	return g.internal.GetMeta().(int)
}
func (g *Node) SetMeta(cnt int) {
	g.internal.SetMeta(cnt)
}
func (g *Node) SetType(t GrammarType) {
	ori := g.internal.GetProperty(Prop)
	ori.Type = t
	g.internal.SetProperty(Prop, ori)
}

func (g *Node) GetGrammar() *Grammar {
	return g.internal.GetProperty(Prop).Gram
}

func (g *Node) AddSymbol(new *Node) int {
	e := g.newEdge(GetEdgeID(g.GetID(), new.GetID()), g, new)
	g.GetGrammar().internal.AddEdge(e)
	return len(g.GetGrammar().internal.GetOutEdges(g.internal)) - 1
}
func getNumber(id string) int {
	ids := strings.Split(id, "#")
	if len(ids) != 2 {
		slog.Info("the id format should be xxx#yyy")
		return 0
	}
	num1, err := strconv.Atoi(ids[1])
	if err != nil {
		slog.Error("strconv atoi", "error", err)
	}
	return num1
}
func (g *Node) GetSymbols() []*Node {
	edges := g.GetGrammar().internal.GetOutEdges(g.internal)
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].GetMeta().(int) > edges[j].GetMeta().(int)
	})
	f := func(edge graph.Edge[string, Property]) *Node {
		return &Node{internal: edge.GetTo()}
	}
	ori := A.Map(f)(edges)
	return ori
}

func (g *Node) GetSymbol(idx int) *Node {
	syms := g.GetSymbols()
	if idx < len(syms) {
		return (syms)[idx]
	}
	return nil
}

func (g *Node) GetContent() string {
	return g.internal.GetProperty(Prop).Content
}
func (g *Node) SetContent(content string) {
	p := g.internal.GetProperty(Prop)
	p.Content = content
	g.internal.SetProperty(Prop, p)
}
func (g *Node) GetDistance() int {
	return g.internal.GetProperty(Prop).DistanceToTerminal
}
func GetEdgeID(father string, child string) string {
	return fmt.Sprintf("%s,%s", father, child)
}

func ExtractEdgeID(id string) (string, string) {
	res := strings.Split(id, ",")
	if len(res) != 2 {
		panic("error in id")
	}
	return res[0], res[1]
}
