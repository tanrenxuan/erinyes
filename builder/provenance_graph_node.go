package builder

import (
	"erinyes/helper"
)

type NodeType int

func (t NodeType) String() string {
	switch t {
	case Process:
		return "Process"
	case File:
		return "File"
	case Socket:
		return "Socket"
	}
	return "Unknown"
}

const (
	Process NodeType = iota
	File
	Socket
)

// ProcessInfo process node's information
type ProcessInfo struct {
	Path          string
	Name          string
	Pid           string
	ContainerID   string
	ContainerName string
	HostID        string
	HostName      string
}

func (p ProcessInfo) Info() string {
	return p.Pid + "_" + p.Name + "#" + p.HostID + "_" + p.ContainerID
}

func (p ProcessInfo) Flag() string {
	return "cluster" + p.HostID + "_" + p.ContainerID
}

func (p ProcessInfo) Shape() string {
	return "box"
}

// FileInfo file node's information
type FileInfo struct {
	Path          string
	ContainerName string
	ContainerID   string
	HostName      string
	HostID        string
}

func (f FileInfo) Info() string {
	return f.Path + "#" + f.HostID + "_" + f.ContainerID
}

func (f FileInfo) Flag() string {
	return "cluster" + f.HostID + "_" + f.ContainerID
}

func (f FileInfo) Shape() string {
	return "ellipse"
}

// SocketInfo socket node's information
type SocketInfo struct {
	DstIP         string
	DstPort       string
	ContainerName string
	ContainerID   string
	HostName      string
	HostID        string
}

func (s SocketInfo) Info() string {
	return s.DstIP + ":" + s.DstPort + "#" + s.HostID + "_" + s.ContainerID
}

func (s SocketInfo) Flag() string {
	return "cluster" + s.HostID + "_" + s.ContainerID
}

func (s SocketInfo) Shape() string {
	return "diamond"
}

type NodeInfo interface {
	Info() string
	Flag() string
	Shape() string
}

// GraphNode is provenance graph node
type GraphNode struct {
	id       int64    // unique node id
	nodeType NodeType // node type
	nodeInfo NodeInfo // node information
}

// ID implements Node interface
func (n GraphNode) ID() int64 {
	return n.id
}

func (n GraphNode) VertexClusterID() string {
	return helper.AddQuotation(n.nodeInfo.Flag())
}

func (n GraphNode) VertexName() string {
	return helper.AddQuotation(n.nodeInfo.Info())
}

func (n GraphNode) VertexShape() string {
	return n.nodeInfo.Shape()
}
