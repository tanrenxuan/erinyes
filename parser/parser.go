package parser

import (
	"bufio"
	"erinyes/logs"
	"os"
)

type Parser interface {
	ParsePushLine(rawLine string) error // 解析原始日志，生成 ParsedLog 放入 Pusher 中
}

type Pusher struct {
	parsedLogCh *chan ParsedLog
}

func (p *Pusher) PushParsedLog(pl ParsedLog) error {
	*p.parsedLogCh <- pl
	return nil
}

// ParseFile 用于解析文件并插入 pusher 中
func ParseFile(name string, parser Parser) error {
	f, err := os.Open(name)
	if err != nil {
		logs.Logger.WithError(err).Errorf("Open file %s failed", name)
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(bufio.NewReader(f))

	for s.Scan() { // 逐行解析 可以考虑并发
		line := s.Text()
		err = parser.ParsePushLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

var SysdigRawChan chan string
var NetRawChan chan string

// ParseSysdigChan 用于实时解析 SysdigRawChan 中的日志并插入 pusher 中
func ParseSysdigChan(parser Parser) {
	SysdigRawChan = make(chan string, 1000)
	for rawString := range SysdigRawChan {
		err := parser.ParsePushLine(rawString)
		if err != nil {
			logs.Logger.Errorf("parse sysdig log failed: %s", rawString)
		}
	}
}

// ParseNetChan 用于实时解析 NetRawChan 中的日志并插入 pusher 中
func ParseNetChan(parser Parser) {
	NetRawChan = make(chan string, 1000)
	for rawString := range NetRawChan {
		err := parser.ParsePushLine(rawString)
		if err != nil {
			logs.Logger.Errorf("parse net log failed: %s", rawString)
		}
	}
}
