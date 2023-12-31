package parser

const (
	SYSDIGTYPE = "sysdig_edge"
	NETTYPE    = "net_edge"

	SOCKETTYPE  = "socket_vertex"
	FILETYPE    = "file_vertex"
	PROCESSTYPE = "process_vertex"
)

type ParsedSysdigLog struct {
	EventCLass string
	Relation   string
	Operation  string
	Time       int64
	UUID       string
}

func (p ParsedSysdigLog) LogType() string {
	return SYSDIGTYPE
}

type ParsedNetLog struct {
	Method     string
	Payload    string
	PayloadLen int
	SeqNum     int
	AckNum     int
	Time       int64
	UUID       string
}

func (p ParsedNetLog) LogType() string {
	return NETTYPE
}

type ProcessVertex struct {
	HostID         string
	HostName       string
	ContainerID    string
	ContainerName  string
	ProcessVPID    string
	ProcessName    string
	ProcessExepath string
}

func (v ProcessVertex) VertexType() string {
	return PROCESSTYPE
}

type FileVertex struct {
	HostID        string
	HostName      string
	ContainerID   string
	ContainerName string
	FilePath      string
}

func (v FileVertex) VertexType() string {
	return FILETYPE
}

type SocketVertex struct {
	HostID        string
	HostName      string
	ContainerID   string
	ContainerName string
	DstIP         string
	DstPort       string
}

func (s SocketVertex) VertexType() string {
	return SOCKETTYPE
}

type ParsedEdge interface {
	LogType() string
}

type ParsedVertex interface {
	VertexType() string
}

type ParsedLog struct {
	Log         ParsedEdge
	StartVertex ParsedVertex
	EndVertex   ParsedVertex
}
