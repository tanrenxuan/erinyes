package models

// DotVertex 抽象出的dot文件顶点
type DotVertex interface {
	VertexClusterID() string
	VertexName() string
	VertexShape() string
}

// DotEdge 抽象出的dot文件边
type DotEdge interface {
	EdgeName() string
	HasEdgeUUID() bool
}
