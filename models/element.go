package models

// DotVertex 抽象出的dot文件顶点
type DotVertex interface {
	VertexClusterID() string // 等于Link的Category
	VertexName() string      // 和LinkID()返回值一样
	VertexShape() string
	LinkID() string
	LinkName() string
	LinkSymbol() string
	LinkInfo() string
	LinkCategory() string
}

// DotEdge 抽象出的dot文件边
type DotEdge interface {
	EdgeName() string
	HasEdgeUUID() bool
	GetUUID() string
	LinkLabel() string
	LinkInfo() string
}
