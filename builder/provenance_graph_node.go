package builder

import "fmt"

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
	return fmt.Sprintf("%s_%s", p.Pid, p.Name)
}

func (p ProcessInfo) Flag() string {
	return p.HostID + "#" + p.ContainerID
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
	return fmt.Sprintf("%s", f.Path)
}

func (f FileInfo) Flag() string {
	return f.HostID + "#" + f.ContainerID
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
	return fmt.Sprintf("%s:%s", s.DstIP, s.DstPort)
}

func (s SocketInfo) Flag() string {
	return s.HostID + "#" + s.ContainerID
}

type NodeInfo interface {
	Info() string
	Flag() string
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

func (n GraphNode) Info() string {
	return fmt.Sprintf("\"%d %s\"", n.ID(), n.nodeInfo.Info())
}

func (n GraphNode) TypeString() string {
	return n.nodeType.String()
}

func (n GraphNode) InfoRaw() NodeInfo {
	return n.nodeInfo
}
