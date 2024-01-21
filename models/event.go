package models

import "erinyes/helper"

type Event struct {
	ID         int    `gorm:"primaryKey;column:id"`
	SrcID      int    `gorm:"column:src_id"`
	DstID      int    `gorm:"column:dst_id"`
	EventClass string `gorm:"column:event_class"`
	Relation   string `gorm:"column:relation"`
	Operation  string `gorm:"column:operation"`
	Time       int64  `gorm:"column:time"`
	UUID       string `gorm:"column:uuid"`
}

func (Event) TableName() string {
	return "event"
}

// EdgeName 实现接口 返回Dot文件中边的名称
func (e Event) EdgeName() string {
	return helper.AddQuotation(e.Relation)
	//return helper.AddQuotation(e.Relation + "_" + e.UUID)
}

func (e Event) HasEdgeUUID() bool {
	if e.UUID != "unknown" {
		return true
	}
	return false
}

func (e Event) GetUUID() string {
	return e.UUID
}
