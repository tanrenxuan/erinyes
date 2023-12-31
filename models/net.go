package models

import "erinyes/helper"

type Net struct {
	ID         int    `gorm:"primaryKey;column:id"`
	SrcID      int    `gorm:"column:src_id"`
	DstID      int    `gorm:"column:dst_id"`
	Method     string `gorm:"column:method"`
	Payload    string `gorm:"column:payload"`
	PayloadLen int    `gorm:"column:payload_len"`
	SeqNum     int    `gorm:"column:seq_num"`
	AckNum     int    `gorm:"column:ack_num"`
	Time       int64  `gorm:"column:time"`
	UUID       string `gorm:"column:uuid"`
}

func (Net) TableName() string {
	return "net"
}

// EdgeName 实现接口 返回Dot文件中边的名称
func (n Net) EdgeName() string {
	return helper.AddQuotation(n.Method)
}
