package builder

import (
	"bytes"
	"erinyes/logs"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"gonum.org/v1/gonum/graph/multi"
	"io/ioutil"
	"os/exec"
)

// entityType2Shape map node type to certain shape
var entityType2Shape = map[NodeType]string{
	Process: "rect", File: "ellipse", Socket: "diamond",
}

// callSystem 执行指定命令
func callSystem(s string, args ...string) error {
	cmd := exec.Command(s, args...)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	fmt.Printf("%s", out.String())
	return err
}

// Visualize 可视化带权有向多重图
func Visualize(g *multi.WeightedDirectedGraph) error {
	graphAst, _ := gographviz.ParseString(`digraph G{}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		return err
	}
	// 填入所有node
	nodes := g.Nodes()
	fmt.Printf("Nodes: %d\n", nodes.Len())
	for nodes.Next() {
		N := nodes.Node()
		n := N.(GraphNode)
		GenerateVertex(n, graph)
	}

	// 填入所有edge
	edges := g.Edges()
	count := 0 // 没有直接计算lines数量的函数，单独计数
	for edges.Next() {
		e := edges.Edge()
		lines := g.WeightedLines(e.From().ID(), e.To().ID())
		for lines.Next() {
			count++
			L := lines.WeightedLine()
			l := L.(GraphLine)
			From := l.From()
			from := From.(GraphNode)
			To := l.To()
			to := To.(GraphNode)
			//if l.TimeStamp != 0 {
			//	if err := graph.AddEdge(from.VertexName(), to.VertexName(), true, map[string]string{"label": fmt.Sprintf("%s_%ds", l.Relation, l.TimeStamp)}); err != nil {
			//		logs.Logger.Warnf("failed to add edge to the graphviz, edge = [from: %s, to: %s]", from.VertexName(), to.VertexName())
			//	}
			//} else {
			//	if err := graph.AddEdge(from.VertexName(), to.VertexName(), true, map[string]string{"label": l.Relation}); err != nil {
			//		logs.Logger.Warnf("failed to add edge to the graphviz, edge = [from: %s, to: %s]", from.VertexName(), to.VertexName())
			//	}
			//}
			if err := graph.AddEdge(from.VertexName(), to.VertexName(), true, map[string]string{"label": l.Relation}); err != nil {
				logs.Logger.Warnf("failed to add edge to the graphviz, edge = [from: %s, to: %s]", from.VertexName(), to.VertexName())
			}
		}
	}
	fmt.Printf("Edges: %d\n", count)
	//fmt.Println(graph.String())
	if err := ioutil.WriteFile("graphs/test.dot", []byte(graph.String()), 0666); err != nil {
		return err
	}
	return callSystem("dot", "-T", "svg", "graphs/test.dot", "-o", "graphs/test.svg")
}
