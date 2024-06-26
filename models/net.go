package models

import (
	"erinyes/helper"
	"fmt"
	"strings"
)

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
	//return helper.AddQuotation(n.Method + "_" + n.UUID)
}
func (n Net) HasEdgeUUID() bool {
	if n.UUID != "" {
		return true
	}
	return false
}

func (n Net) GetUUID() string {
	return n.UUID
}

func (n Net) LinkLabel() string {
	return strings.ToLower(n.Method)
}

func (n Net) LinkInfo() string {
	return fmt.Sprintf("method:%s\npayload_len:%d\nseq_num:%d\nack_num:%d\ntime:%d\nuuid:%s", n.Method, n.PayloadLen, n.SeqNum, n.AckNum, n.Time, n.UUID)
}
