package schemas

import (
	"fmt"
	"github.com/lucasjones/reggen"
	"strings"
)

type Derivation struct {
	*Grammar
	EdgeHistory []string
	SymbolCnt   map[string]int // 为了给语法图上的Node做标记
}

func (d *Derivation) getNodeID(id string) string {
	//return id
	return fmt.Sprintf("%s#%d", id, d.SymbolCnt[id])
}

// AddNode convention: When adding a Edge, by AddEdge, we first add Node to the graph
func (d *Derivation) AddNode(node *Node) {
	if d.internal.GetVertexById(d.getNodeID(node.GetID())) != nil { // already exists
		d.SymbolCnt[node.GetID()]++
	}
	newnode := node.Clone(d.Grammar)
	newnode.SetID(d.getNodeID(node.GetID()))
}

// AddEdge convention: When adding a Edge, by AddEdge, we first AddNode to the graph
func (d *Derivation) AddEdge(from, to *Node) {
	newfrom := from.Clone(d.Grammar)
	newto := to.Clone(d.Grammar)

	newfrom.SetID(d.getNodeID(from.GetID()))
	newto.SetID(d.getNodeID(to.GetID()))
	from.SetMeta(from.GetMeta() + 1)

	newfrom.AddSymbol(newto)
	d.EdgeHistory = append(d.EdgeHistory, GetEdgeID(newfrom.GetID(), newto.GetID()))
}
func isTermPreserve(content string) bool {
	return (content[0] == content[len(content)-1]) && ((content[0] == '\'') || content[0] == '"')
}

func (d *Derivation) GetResult(custom func(content string) string) string {
	root := d.Grammar.GetNode(d.Grammar.GetStartSym() + "#0")
	if root == nil {
		return ""
	}
	res := ""
	d.SymbolCnt = make(map[string]int)

	dfs(root, func(cur *Node) {
		if cur.GetType() == GrammarTerminal {
			content := cur.GetContent()
			if isTermPreserve(content) {
				tmp := strings.Trim(content, "'\"")
				switch content[0] {
				case '"':
					var err error
					content, err = reggen.Generate(tmp, 10)
					if err != nil {
						panic(err)
					}
				case '\'':
					content = tmp
				default:
					panic("error in generating terminal")
				}
			}
			if custom != nil {
				content = custom(content)
			}
			res += content
			d.SymbolCnt[cur.GetID()]++
		}
	})
	return res
}

func (d *Derivation) Visualize() {

}
