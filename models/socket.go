package models

import (
	"erinyes/conf"
	"erinyes/helper"
	"erinyes/logs"
	"fmt"
	"gorm.io/gorm"
)

type Socket struct {
	ID            int    `gorm:"primaryKey;column:id"`
	HostID        string `gorm:"column:host_id"`
	HostName      string `gorm:"column:host_name"`
	ContainerID   string `gorm:"column:container_id"`
	ContainerName string `gorm:"column:container_name"`
	DstIP         string `gorm:"column:dst_ip"`
	DstPort       string `gorm:"column:dst_port"`
}

func (Socket) TableName() string {
	return "socket"
}

func (s *Socket) FindByID(db *gorm.DB, id int) bool {
	err := db.First(s, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Logger.Errorf("can't find socket by id = %d", id)
		} else {
			logs.Logger.Errorf("query socket by id = %d failed: %w", id, err)
		}
		return false
	}
	return true
}

// VertexClusterID 实现点的接口，返回dot文件中点的唯一标识
func (s Socket) VertexClusterID() string {
	return helper.AddQuotation("cluster" + s.HostID + "_" + s.ContainerID)
}

// VertexName 返回该节点在dot文件中的名称
func (s Socket) VertexName() string {
	//if len(s.DstPort) >= 5 { // 减少图中 socket 的数量（边的数量不变，聚合到了一个socket），但是db中的图结构不变
	//	return helper.AddQuotation(s.DstIP + ":" + "10000" + "#" + s.HostID + "_" + s.ContainerID)
	//}
	return helper.AddQuotation(s.DstIP + ":" + s.DstPort + "#" + s.HostID + "_" + s.ContainerID)
}

// VertexShape 返回该节点的形状
func (s Socket) VertexShape() string {
	return "diamond"
}

// RelateHostAndCin 关联主机IP和Cin0的IP
func (s *Socket) RelateHostAndCin() {
	if s.DstIP == conf.Config.Cin0IP {
		s.DstIP = conf.Config.HostIP
		s.DstPort = "8085"
	} else if s.DstIP == conf.Config.HostIP {
		s.DstPort = "8085"
	} else if s.DstIP == "127.0.0.1" { // 只会修改流量日志里的socket，因为审计日志中全部修改为了localhost
		s.DstIP = conf.Config.HostIP
		s.DstPort = "8085"
	}
}

// UnionGateway 统一gateway
func (s *Socket) UnionGateway() {
	gateways := conf.Config.GatewayMap
	if _, exist := gateways[s.DstIP]; exist { // 该socket是gateway
		s.DstIP = "gateway"
		s.DstPort = "8080"
	}
}

func (s Socket) LinkID() string {
	return s.DstIP + ":" + s.DstPort + "#" + s.HostID + "_" + s.ContainerID
}

func (s Socket) LinkName() string {
	return s.DstIP + ":" + s.DstPort
}

func (s Socket) LinkSymbol() string {
	return "diamond"
}

func (s Socket) LinkInfo() string {
	return fmt.Sprintf("host_id:%s\ncontainer_id:%s\nip:%s\nport:%s", s.HostID, s.ContainerID, s.DstIP, s.DstPort)
}

func (s Socket) LinkCategory() string {
	return s.HostID + "_" + s.ContainerID
}
