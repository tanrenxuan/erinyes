package parser

import (
	"encoding/json"
	"erinyes/logs"
	"regexp"
	"strconv"
	"strings"
)

type NetJson struct {
	IPSrc      string  `json:"ip_src"`
	PortSrc    int     `json:"port_src"`
	IPDst      string  `json:"ip_dst"`
	PortDst    int     `json:"port_dst"`
	SeqNum     int     `json:"sequence_num"`
	AckNum     int     `json:"acknowledge_num"`
	PayLoadLen int     `json:"payload_len"`
	PayLoad    string  `json:"payload"`
	TimeStamp  float64 `json:"time_stamp"`
}

type NetLog struct {
	IPSrc      string
	PortSrc    string
	IPDst      string
	PortDst    string
	SeqNum     int
	AckNum     int
	PayLoadLen int
	Method     string
	Time       int64
	UUID       string
}

// SplitNetLine 解析原始流量日志
func SplitNetLine(rawLine string) (error, *NetLog) {
	var netJson NetJson
	err := json.Unmarshal([]byte(rawLine), &netJson)
	if err != nil {
		logs.Logger.WithError(err).Errorf("解析JSON时发生错误")
		return err, nil
	}
	// 从payload中解析request or response
	method := "UNKNOWN"
	index := strings.Index(netJson.PayLoad, " ")
	if index != -1 {
		firstStr := netJson.PayLoad[:index]
		if strings.HasPrefix(firstStr, "HTTP") { // response
			method = "RESPONSE"
		} else if firstStr == "GET" {
			method = "GET"
		} else if firstStr == "POST" {
			method = "POST"
		} else if firstStr == "DELETE" {
			method = "DELETE"
		} else if firstStr == "PUT" {
			method = "PUT"
		}
	}
	// 从payload中解析出uuid
	var uuid string
	uuidRegex := regexp.MustCompile(`uuid: (\d+)`)
	matches := uuidRegex.FindStringSubmatch(netJson.PayLoad)
	if len(matches) > 1 {
		uuid = matches[1]
	}
	netData := NetLog{
		IPSrc:      netJson.IPSrc,
		PortSrc:    strconv.Itoa(netJson.PortSrc),
		IPDst:      netJson.IPDst,
		PortDst:    strconv.Itoa(netJson.PortDst),
		SeqNum:     netJson.SeqNum,
		AckNum:     netJson.AckNum,
		PayLoadLen: netJson.PayLoadLen,
		Method:     method,
		Time:       int64(netJson.TimeStamp * 1000000), // 16位，微秒级别
		UUID:       uuid,
	}
	return nil, &netData
}
