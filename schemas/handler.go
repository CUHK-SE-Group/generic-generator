package schemas

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

const (
	CatHandlerName      = "cat_handler"
	OrHandlerName       = "or_handler"
	IDHandlerName       = "id_handler"
	BracketHandlerName  = "bracket_handler"
	PlusHandlerName     = "plus_handler"
	TerminalHandlerName = "terminal_handler"
	SubHandlerName      = "sub_handler"
	RepHandlerName      = "rep_handler"
	TraceHandlerName    = "trace_handler"
	OptionHandlerName   = "option_handler"
)

type Handler interface {
	Handle(*Chain, *Context, ResponseCallBack)
	HookRoute() []regexp.Regexp
	Name() string
	Type() GrammarType
}

// CatHandler default implementation for Catenation
type CatHandler struct {
}

func (h *CatHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	if len(ctx.CurrentNode.GetSymbols()) == 0 {
		chain.Next(ctx, cb)
		return
	}
	for i := len(ctx.CurrentNode.GetSymbols()) - 1; i >= 0; i-- {
		ctx.ResultBuffer = append(ctx.ResultBuffer, ctx.CurrentNode.GetSymbols()[i])
	}
	//for i := 0; i < len(cur.GetSymbols()); i++ {
	//	ctx.Result.AddNode((cur.GetSymbols())[i])
	//	ctx.Result.AddEdge(cur, (cur.GetSymbols())[i])
	//}
	chain.Next(ctx, cb)
}

func (h *CatHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *CatHandler) Name() string {
	return CatHandlerName
}

func (h *CatHandler) Type() GrammarType {
	return GrammarProduction | GrammarCatenate
}

// OrHandler default implementation of Or, randomly choose a child to generate
type OrHandler struct {
}

func (h *OrHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	if len(ctx.CurrentNode.GetSymbols()) == 0 {
		chain.Next(ctx, cb)
		return
	}
	idx := rand.Int() % len(ctx.CurrentNode.GetSymbols())
	ctx.ResultBuffer = append(ctx.ResultBuffer, ctx.CurrentNode.GetSymbol(idx))

	//ctx.Result.AddNode((cur.GetSymbols())[idx])
	//ctx.Result.AddEdge(cur, (cur.GetSymbols())[idx])
	//ctx.VisitedEdge[GetEdgeID(cur.GetID(), (cur.GetSymbols())[idx].GetID())]++
	chain.Next(ctx, cb)
}

func (h *OrHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *OrHandler) Name() string {
	return OrHandlerName
}

func (h *OrHandler) Type() GrammarType {
	return GrammarOR
}

type IDHandler struct {
}

func (h *IDHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	// 是Identifier, 那么去找新的production
	node := ctx.Grammar.GetNode(ctx.CurrentNode.GetContent())
	if node == nil {
		slog.Error("The identifier does not Existed", "id", ctx.CurrentNode.GetContent())
		return // omit error
	}
	//ctx.Result.AddNode(node)
	//ctx.Result.AddEdge(ctx.CurrentNode, node)
	ctx.ResultBuffer = append(ctx.ResultBuffer, node)
	chain.Next(ctx, cb)
}

func (h *IDHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *IDHandler) Name() string {
	return IDHandlerName
}

func (h *IDHandler) Type() GrammarType {
	return GrammarID
}

type RepHandler struct {
}

func (r *RepHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	// 默认设置 10% 的概率来重复一次
	if rand.Intn(10) > 8 {
		ctx.ResultBuffer = append(ctx.ResultBuffer, ctx.CurrentNode.GetSymbols()...)
		//for _, node := range ctx.CurrentNode.GetSymbols() {
		//	ctx.Result.AddNode(node)
		//	ctx.Result.AddEdge(cur, node)
		//}
	}
	chain.Next(ctx, cb)
}

func (r *RepHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (r *RepHandler) Name() string {
	return RepHandlerName
}

func (r *RepHandler) Type() GrammarType {
	return GrammarREP
}

type TermHandler struct {
}

func (h *TermHandler) isTermPreserve(g *Node) bool {
	content := g.GetContent()
	return (content[0] == content[len(content)-1]) && ((content[0] == '\'') || content[0] == '"')
}

func (h *TermHandler) stripQuote(content string) string {
	if content[0] == content[len(content)-1] {
		if (content[0] == '\'') || (content[0] == '"') {
			return content[1 : len(content)-1]
		}
	}
	return content
}

func (h *TermHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	if len(ctx.CurrentNode.GetSymbols()) != 0 {
		slog.Error("Pattern mismatched[Terminal]")
		return
	}
	if len(ctx.tmp) == 0 {
		ctx.tmp = make([]string, 0)
	}
	ctx.tmp = append(ctx.tmp, strings.Trim(ctx.CurrentNode.GetContent(), "'"))
	fmt.Printf("============correct: ")
	for i := len(ctx.tmp) - 1; i >= 0; i-- {
		fmt.Printf("%s", ctx.tmp[i])
	}
	fmt.Println()

	//ctx.Result.AddEdge(cur, cur) // 用一个自环标记到达了最后的终结符节点
	chain.Next(ctx, cb)
}

func (h *TermHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *TermHandler) Name() string {
	return TerminalHandlerName
}

func (h *TermHandler) Type() GrammarType {
	return GrammarTerminal
}

type BracketHandler struct {
}

func (h *BracketHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	//cur := ctx.SymbolStack.Top()
	//ctx.SymbolStack.Pop()
	children := ctx.CurrentNode.GetSymbols()
	if len(children) == 0 {
		slog.Error("Pattern mismatched[Identifier]")
		return
	}
	// todo, 注释这段代码。这段代码是为了测试
	if strings.Contains(ctx.CurrentNode.GetContent(), "SP") {
		for i := len(children) - 1; i >= 0; i-- {
			//ctx.SymbolStack.Push(children[i])
			ctx.ResultBuffer = append(ctx.ResultBuffer, children[i])
		}
		//for i := 0; i < len(children); i++ {
		//	ctx.Result.AddNode(children[i])
		//	ctx.Result.AddEdge(ctx.CurrentNode, children[i])
		//}
	}
	chain.Next(ctx, cb)
}

func (h *BracketHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *BracketHandler) Name() string {
	return BracketHandlerName
}

func (h *BracketHandler) Type() GrammarType {
	return GrammarOptional
}

type PlusHandler struct {
}

func (h *PlusHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	//cur := ctx.SymbolStack.Top()
	//ctx.SymbolStack.Pop()
	children := ctx.CurrentNode.GetSymbols()
	if len(children) == 0 {
		slog.Error("Pattern mismatched[Identifier]")
		return
	}
	for j := 0; j < rand.Intn(10)+1; j++ {
		for i := len(children) - 1; i >= 0; i-- {
			ctx.ResultBuffer = append(ctx.ResultBuffer, children[i])
		}
	}

	chain.Next(ctx, cb)

}

func (h *PlusHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *PlusHandler) Name() string {
	return PlusHandlerName
}

func (h *PlusHandler) Type() GrammarType {
	return GrammarPLUS
}

type SubHandler struct {
}

func (h *SubHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	chain.Next(ctx, cb)
}

func (h *SubHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *SubHandler) Name() string {
	return SubHandlerName
}

func (h *SubHandler) Type() GrammarType {
	return GrammarSUB
}

type TraceHandler struct {
}

func (h *TraceHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	chain.Next(ctx, cb)
}

func (h *TraceHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *TraceHandler) Name() string {
	return TraceHandlerName
}

func (h *TraceHandler) Type() GrammarType {
	return math.MaxInt
}

type OptionHandler struct {
}

func (h *OptionHandler) Handle(chain *Chain, ctx *Context, cb ResponseCallBack) {
	fmt.Println("dealing with", ctx.CurrentNode.GetContent())
	chain.Next(ctx, cb)
}

func (h *OptionHandler) HookRoute() []regexp.Regexp {
	return make([]regexp.Regexp, 0)
}

func (h *OptionHandler) Name() string {
	return OptionHandlerName
}

func (h *OptionHandler) Type() GrammarType {
	return GrammarChoice
}
