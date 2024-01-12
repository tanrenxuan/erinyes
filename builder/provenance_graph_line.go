package builder

import "gonum.org/v1/gonum/graph"

type GraphLine struct {
	F, T      graph.Node
	W         float64 // 稀有路径得分
	Relation  string  // 边的关系
	TimeStamp int64   // 17位时间戳
	UID       int64   // 两个共同顶点之间的平行边，需要用UID区分
}

// From To ReversedLine ID Weight implements the WeightedLine interface
func (l GraphLine) From() graph.Node { return l.F }

func (l GraphLine) To() graph.Node { return l.T }

func (l GraphLine) ReversedLine() graph.Line { l.F, l.T = l.T, l.F; return l }

func (l GraphLine) ID() int64 { return l.UID }

// Weight returns the weight of the edge.
func (l GraphLine) Weight() float64 { return l.W }
