package models

import (
	"erinyes/helper"
	"erinyes/logs"
	"fmt"
	"gorm.io/gorm"
)

type File struct {
	ID            int    `gorm:"primaryKey;column:id"`
	HostID        string `gorm:"column:host_id"`
	HostName      string `gorm:"column:host_name"`
	ContainerID   string `gorm:"column:container_id"`
	ContainerName string `gorm:"column:container_name"`
	FilePath      string `gorm:"column:file_path"`
}

func (File) TableName() string {
	return "file"
}

func (f *File) FindByID(db *gorm.DB, id int) bool {
	err := db.First(f, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Logger.Errorf("can't find file by id = %d", id)
		} else {
			logs.Logger.Errorf("query file by id = %d failed: %w", id, err)
		}
		return false
	}
	return true
}

// VertexClusterID 实现点的接口，返回dot文件中点的唯一标识
func (f File) VertexClusterID() string {
	return helper.AddQuotation("cluster" + f.HostID + "_" + f.ContainerID)
}

// VertexName 返回该节点在dot文件中的名称
func (f File) VertexName() string {
	return helper.AddQuotation(f.FilePath + "#" + f.HostID + "_" + f.ContainerID)
}

// VertexShape 返回该节点的形状
func (f File) VertexShape() string {
	return "ellipse"
}

// LinkID 节点的唯一标识（全局唯一）
func (f File) LinkID() string {
	return f.FilePath + "#" + f.HostID + "_" + f.ContainerID
}

func (f File) LinkName() string {
	return f.FilePath
}

func (f File) LinkSymbol() string {
	return "circle"
}

func (f File) LinkInfo() string {
	return fmt.Sprintf("host_id:%s\ncontainer_id:%s\nfile_path:%s", f.HostID, f.ContainerID, f.FilePath)
}

func (f File) LinkCategory() string {
	return f.HostID + "_" + f.ContainerID
}
