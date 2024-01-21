package parser

import (
	"erinyes/conf"
	"erinyes/logs"
	"strings"
)

type NetParser struct {
	pusher *Pusher
}

func NewNetParser(pusher *Pusher) *NetParser {
	return &NetParser{
		pusher: pusher,
	}
}

func (p *NetParser) ParserType() string {
	return NET
}

// ParsePushLine 实现 parser 的接口
func (p *NetParser) ParsePushLine(rawLine string) error {
	err, netLog := SplitNetLine(rawLine)
	if err != nil {
		return err
	}
	// alastor 会判断 IP 是否为 function 的 ip，则另一个 ip 是 gateway
	// erinyes 记录的网络日志中，除了gateway、function 的 ip，还有很多其他的，因此如实记录各个ip即可
	pl := ParsedLog{}
	if containerNameAndID, ok := conf.Config.IPMap[netLog.IPSrc]; ok { // 判断 src 是否为几个 function 中的 ip
		result := strings.Split(containerNameAndID, "$")
		//pl.StartVertex = SocketVertex{
		//	HostID:        conf.MockHostID,
		//	HostName:      conf.MockHostName,
		//	ContainerID:   result[1],
		//	ContainerName: result[0],
		//	DstIP:         netLog.IPSrc,
		//	DstPort:       netLog.PortSrc,
		//}
		pl.StartVertex = ProcessVertex{
			HostID:         conf.MockHostID,
			HostName:       conf.MockHostName,
			ContainerID:    result[1],
			ContainerName:  result[0],
			ProcessVPID:    "1",
			ProcessName:    "fwatchdog",
			ProcessExepath: "unknwon",
		}
	} else {
		pl.StartVertex = SocketVertex{
			HostID:        conf.MockHostID,
			HostName:      conf.MockHostName,
			ContainerID:   conf.OuterContainerID,
			ContainerName: conf.OuterContainerName,
			DstIP:         netLog.IPSrc,
			DstPort:       netLog.PortSrc,
		}
	}
	if containerNameAndID, ok := conf.Config.IPMap[netLog.IPDst]; ok {
		result := strings.Split(containerNameAndID, "$")
		//pl.EndVertex = SocketVertex{
		//	HostID:        conf.MockHostID,
		//	HostName:      conf.MockHostName,
		//	ContainerID:   result[1],
		//	ContainerName: result[0],
		//	DstIP:         netLog.IPDst,
		//	DstPort:       netLog.PortDst,
		//}
		pl.EndVertex = ProcessVertex{
			HostID:         conf.MockHostID,
			HostName:       conf.MockHostName,
			ContainerID:    result[1],
			ContainerName:  result[0],
			ProcessVPID:    "1",
			ProcessName:    "fwatchdog",
			ProcessExepath: "unknwon",
		}
	} else {
		pl.EndVertex = SocketVertex{
			HostID:        conf.MockHostID,
			HostName:      conf.MockHostName,
			ContainerID:   conf.OuterContainerID,
			ContainerName: conf.OuterContainerName,
			DstIP:         netLog.IPDst,
			DstPort:       netLog.PortDst,
		}
	}

	if pl.StartVertex.VertexType() == SOCKETTYPE && pl.EndVertex.VertexType() == SOCKETTYPE {
		pl.Log = ParsedNetLog{
			Method:     netLog.Method,
			PayloadLen: netLog.PayLoadLen,
			SeqNum:     netLog.SeqNum,
			AckNum:     netLog.AckNum,
			Time:       netLog.Time,
			UUID:       netLog.UUID,
		}
		p.pusher.PushParsedLog(pl)
	} else if pl.StartVertex.VertexType() == PROCESSTYPE && pl.EndVertex.VertexType() == SOCKETTYPE {
		pl.Log = ParsedSysdigLog{
			EventCLass: NETWORKV1,
			Relation:   netLog.Method,
			Operation:  netLog.Method,
			Time:       netLog.Time,
			UUID:       netLog.UUID,
		}
		p.pusher.PushParsedLog(pl)
	} else if pl.StartVertex.VertexType() == SOCKETTYPE && pl.EndVertex.VertexType() == PROCESSTYPE {
		pl.Log = ParsedSysdigLog{
			EventCLass: NETWORKV2,
			Relation:   netLog.Method,
			Operation:  netLog.Method,
			Time:       netLog.Time,
			UUID:       netLog.UUID,
		}
		p.pusher.PushParsedLog(pl)
	} else {
		logs.Logger.Error("There are two process vertex in on edge")
	}
	return nil
}
