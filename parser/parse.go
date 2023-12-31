package parser

import (
	"erinyes/logs"
	"sync"
)

var wgParser = sync.WaitGroup{}
var wgInserter = sync.WaitGroup{}

// FileLogParse 用来解析 sysdig 日志和流量日志
func FileLogParse(repeat bool, sysdigFilepath string, netFilepath string) {
	pChan := make(chan ParsedLog, 1000)
	inserter := Inserter{ParsedLogCh: &pChan}
	// 并发解析日志并插入数据库
	concurrencyNum := 10
	for idx := 0; idx < concurrencyNum; idx++ {
		wgInserter.Add(1)
		idx := idx
		go func() {
			defer wgInserter.Done()
			inserter.Insert(idx, repeat)
		}()
	}
	if sysdigFilepath != "" {
		addFileLogParse(NewSysdigParser(&Pusher{&pChan}), sysdigFilepath)
	}
	if netFilepath != "" {
		addFileLogParse(NewNetParser(&Pusher{&pChan}), netFilepath)
	}
	wgParser.Wait()
	close(pChan)
	wgInserter.Wait()
}

// addFileLogParse 新增日志解析器，从指定文件中将原始日志解析为ParsedLog
func addFileLogParse(_parser Parser, filename string) {
	wgParser.Add(1)
	go func() {
		defer wgParser.Done()
		parser := _parser
		err := ParseFile(filename, parser)
		if err != nil {
			logs.Logger.WithError(err).Fatalf("Parse %s failed", filename)
		}
	}()
}
