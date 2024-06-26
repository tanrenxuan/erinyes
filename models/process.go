package models

import (
	"erinyes/helper"
	"erinyes/logs"
	"fmt"
	"gorm.io/gorm"
)

type Process struct {
	ID             int    `gorm:"primaryKey;column:id"`
	HostID         string `gorm:"column:host_id"`
	HostName       string `gorm:"column:host_name"`
	ContainerID    string `gorm:"column:container_id"`
	ContainerName  string `gorm:"column:container_name"`
	ProcessVPID    string `gorm:"column:process_vpid"`
	ProcessName    string `gorm:"column:process_name"`
	ProcessExepath string `gorm:"column:process_exe_path"`
}

func (Process) TableName() string {
	return "process"
}

func (p *Process) FindByID(db *gorm.DB, id int) bool {
	err := db.First(p, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Logger.Errorf("can't find process by id = %d", id)
		} else {
			logs.Logger.Errorf("query process by id = %d failed: %w", id, err)
		}
		return false
	}
	return true
}

// VertexClusterID 实现点的接口，返回dot文件中点的唯一标识
func (p Process) VertexClusterID() string {
	return helper.AddQuotation("cluster" + p.HostID + "_" + p.ContainerID)
}

// VertexName 返回该节点在dot文件中的名称
func (p Process) VertexName() string {
	return helper.AddQuotation(p.ProcessVPID + "_" + p.ProcessName + "#" + p.HostID + "_" + p.ContainerID)
}

// VertexShape 返回该节点的形状
func (p Process) VertexShape() string {
	return "box"
}

func (p Process) LinkID() string {
	return p.ProcessVPID + "_" + p.ProcessName + "#" + p.HostID + "_" + p.ContainerID
}

func (p Process) LinkName() string {
	return p.ProcessVPID + "_" + p.ProcessName
}

func (p Process) LinkSymbol() string {
	return "rect"
}

func (p Process) LinkInfo() string {
	return fmt.Sprintf("host_id:%s\ncontainer_id:%s\nprocess_vpid:%s\nprocess_name:%s\nprocess_exe_path:%s", p.HostID, p.ContainerID, p.ProcessVPID, p.ProcessName, p.ProcessExepath)
}

func (p Process) LinkCategory() string {
	return p.HostID + "_" + p.ContainerID
}
