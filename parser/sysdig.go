package parser

import (
	"erinyes/conf"
	"erinyes/helper"
	"erinyes/logs"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	NASTR     string = "<NA>"
	NILSTR    string = ""
	PROCESS   string = "Process"    // process -> process
	NETWORKV1 string = "Network_V1" // process -> socket
	NETWORKV2 string = "Network_V2" // socket -> process
	FILEV1    string = "File_V1"    // process -> file
	FILEV2    string = "File_V2"    // file -> process
)

// syscall
const (
	// process
	SYS_FORK   string = "fork"
	SYS_VFORK  string = "vfork"
	SYS_CLONE  string = "clone"
	SYS_EXECVE string = "execve"

	// network
	// server
	SYS_BIND    string = "bind"
	SYS_LISTEN  string = "listen"
	SYS_ACCEPT  string = "accept"
	SYS_ACCEPT4 string = "accept4"
	// client
	SYS_CONNECT string = "connect"
	// communication
	SYS_SENDTO   string = "sendto"
	SYS_RECVFROM string = "recvfrom"

	// file
	SYS_OPEN   string = "open"
	SYS_OPENAT string = "openat"
	SYS_READ   string = "read"
	SYS_READV  string = "readv"
	SYS_WRITE  string = "write"
	SYS_WRITEV string = "writev"
)

const UNKNOWN string = "unknown"

type SysdigLog struct {
	Time          int64
	Tid           string // thread id
	ProcessName   string
	Pid           string
	VPid          string // pid in container
	Dir           string // direction
	EventType     string // syscall
	Fd            string
	PPid          string
	Cmd           string
	Ret           string // syscall return
	ContainerID   string
	ContainerName string
	Info          []string // syscall parameters
	HostID        string
	HostName      string
}

// Convert2Timestamp 解析原始日期+时间字符串，转换为微秒级16位时间戳
func Convert2Timestamp(timeStr string) (int64, error) {
	layout := "2006-01-02 15:04:05.999999999" // 输入时间的格式
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return 0, err
	}
	t, err := time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		return 0, err
	}
	return t.UnixNano() / int64(time.Microsecond), nil
}

// Convert2Datetime 解析16位时间戳为日期字符串，精确到毫秒
func Convert2Datetime(timestamp int64) (string, error) {
	timestamp /= 1000 // 改为毫秒级字符串
	timeObj := time.Unix(0, timestamp*int64(time.Millisecond))
	//timeObj := time.Unix(0, timestamp*int64(time.Microsecond)) // 微秒级字符串
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "", err
	}
	timeInTargetZone := timeObj.In(loc)

	formattedTime := timeInTargetZone.Format("2006-01-02 15:04:05.999999999")
	return formattedTime, nil
}

func SplitSysdigLine(rawLine string) (error, *SysdigLog) {
	fields := strings.Split(rawLine, " ") // args 在最后
	if len(fields) < 15 {
		return fmt.Errorf("not enough fileds"), nil
	}
	timestamp, err := Convert2Timestamp(fields[0] + " " + fields[1])
	if err != nil {
		logs.Logger.Errorf("Parse datetime to timestamp failed, datetime is %s", fields[0]+" "+fields[1])
		return err, nil
	}
	return nil, &SysdigLog{
		Time:          timestamp,
		ProcessName:   fields[2],
		Tid:           fields[3],
		Pid:           fields[4],
		VPid:          fields[5],
		Dir:           fields[6],
		EventType:     fields[7],
		Fd:            fields[8],
		PPid:          fields[9],
		Cmd:           fields[10],
		Ret:           fields[11],
		ContainerID:   fields[12],
		ContainerName: fields[13],
		Info:          fields[14:],
		HostID:        conf.MockHostID, // 非多主机场景下直接mock
		HostName:      conf.MockHostName,
	}
}

func (s *SysdigLog) IsProcessCall() bool {
	if s.EventType == SYS_CLONE ||
		s.EventType == SYS_FORK ||
		s.EventType == SYS_VFORK ||
		s.EventType == SYS_EXECVE {
		return true
	}
	return false
}

// IsSocket 判断 fd 是否为 ip:port->ip:port 格式
func IsSocket(fd string) bool {
	socketRegex := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+:\d+->\d+\.\d+\.\d+\.\d+:\d+$`)
	return socketRegex.MatchString(fd)
}

func (s *SysdigLog) IsNetCall() bool {
	if s.EventType == SYS_BIND ||
		s.EventType == SYS_LISTEN ||
		s.EventType == SYS_ACCEPT ||
		s.EventType == SYS_ACCEPT4 ||
		s.EventType == SYS_CONNECT ||
		s.EventType == SYS_SENDTO ||
		s.EventType == SYS_RECVFROM {
		return true
	} else if s.EventType == SYS_READ && IsSocket(s.Fd) ||
		s.EventType == SYS_WRITE && IsSocket(s.Fd) {
		return true
	}
	return false
}

func (s *SysdigLog) IsFileCall() bool {
	if s.EventType == SYS_WRITE ||
		s.EventType == SYS_WRITEV ||
		s.EventType == SYS_READ ||
		s.EventType == SYS_READV ||
		s.EventType == SYS_OPEN ||
		s.EventType == SYS_OPENAT {
		return true
	}
	return false
}

// MustExtractFourTuple 强制解析出四元组 (src_ip,src_port,dst_ip,dst_port)
func (s *SysdigLog) MustExtractFourTuple() (string, string, string, string) {
	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+):(\d+)->(\d+\.\d+\.\d+\.\d+):(\d+)`)
	matches := re.FindStringSubmatch(s.Fd)
	leftIP, leftPort, rightIP, rightPort := matches[1], matches[2], matches[3], matches[4]
	// 实验中发现当容器与dns服务器通信时，ip是宿主机ip，并非容器ip
	if leftPort == "53" {
		return rightIP, rightPort, leftIP, leftPort
	}
	if rightPort == "53" {
		return leftIP, leftPort, rightIP, rightPort
	}

	if _, ok := conf.Config.IPMap[leftIP]; ok {
		return leftIP, leftPort, rightIP, rightPort
	}
	if _, ok := conf.Config.IPMap[rightIP]; ok {
		return rightIP, rightPort, leftIP, leftPort
	}

	// 均不是容器的 ip
	if (leftIP == "localhost" || leftIP == "127.0.0.1") && (rightIP == "localhost" || rightIP == "127.0.0.1") {
		return "localhost", leftPort, "localhost", rightPort
	}

	logs.Logger.Warn(s.Fd + " needs to be checked") // 理论上不应该走到这一步，需要实验中继续观察日志
	return leftIP, leftPort, rightIP, rightPort
}

// ExtractPort 根据 Fd 解析 port，可能回解析失败
func (s *SysdigLog) ExtractPort() (string, bool) {
	portRegex := regexp.MustCompile(`:::(\d+)`)
	matches := portRegex.FindStringSubmatch(s.Fd)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

// FilteredFilePath 聚合一部分路径前缀，即前缀相同的节点看作同一个，否则 file 节点太多
func (s *SysdigLog) FilteredFilePath() string {
	if strings.HasPrefix(s.Fd, "/home/app/node_modules") {
		return "/home/app/node_modules"
	}
	return s.Fd
}

// IsNodeTriggerStartLog 判断是否为 node 插入的开始分割日志
func (s *SysdigLog) IsNodeTriggerStartLog() bool {
	if len(s.Info) < 4 { // e.g. res=77 data=flag_data is uuid
		return false
	}
	if s.ProcessName == "node" && s.EventType == SYS_WRITE && s.Info[1] == "data=flag_data" {
		return true
	}
	return false
}

// IsNodeTriggerEndLog 判断是否为 node 插入的结束分割日志
func (s *SysdigLog) IsNodeTriggerEndLog() bool {
	if len(s.Info) < 4 { // e.g. res=77 data=end_flag_data is uuid
		return false
	}
	if s.ProcessName == "node" && s.EventType == SYS_WRITE && s.Info[1] == "data=end_flag_data" {
		return true
	}
	return false
}

// IsOfwatchdogTriggerStartLog 判断是否为 ofwatchdog 插入的开始分割日志
func (s *SysdigLog) IsOfwatchdogTriggerStartLog() bool {
	if len(s.Info) < 4 { // e.g. res=62 data=start_ofwatchdog_flag_data is uuid
		return false
	}
	if s.ProcessName == "fwatchdog" && s.EventType == SYS_WRITE && s.Info[1] == "data=start_ofwatchdog_flag_data" {
		return true
	}
	return false
}

// IsOfwatchdogTriggerEndLog 判断是否为 ofwatchdog 插入的结束分割日志
func (s *SysdigLog) IsOfwatchdogTriggerEndLog() bool {
	if len(s.Info) < 4 { // e.g. res=62 data=start_ofwatchdog_flag_data is uuid
		return false
	}
	if s.ProcessName == "fwatchdog" && s.EventType == SYS_WRITE && s.Info[1] == "data=end_ofwatchdog_flag_data" {
		return true
	}
	return false
}

// ConvertSysOpen 将 open 系统调用转换为 read 或 write
func (s *SysdigLog) ConvertSysOpen() (error, string) {
	if s.EventType != SYS_OPEN && s.EventType != SYS_OPENAT {
		return fmt.Errorf("not open syscall"), ""
	}
	flag := strings.Join(s.Info, " ")
	re := regexp.MustCompile(`flags=\d+\(([^)]+)\)`)
	match := re.FindStringSubmatch(flag) // match[0]是整个字符串 match[1]是匹配字符串
	if len(match) < 2 {
		return fmt.Errorf("can't parse flag mode: %s", flag), ""
	}
	params := strings.Split(match[1], "|")
	for i, param := range params {
		params[i] = strings.TrimSpace(param)
	}
	// 判断 该 open 是 read 还是 write，参考alastor
	set := make(map[string]bool)
	var result []string
	for _, v := range params {
		set[v] = true
	}
	writeParams := []string{"O_WRONLY", "O_RDWR", "O_APPEND", "O_CREAT", "O_TRUNC"}
	for _, v := range writeParams {
		if set[v] {
			result = append(result, v)
		}
	}
	if len(result) > 0 { // write
		return nil, SYS_WRITE
	} else if set["O_RDONLY"] { // read
		return nil, SYS_READ
	}
	// 该 open 无法转换，忽略即可
	return fmt.Errorf("can't convert this open syscall"), ""
}

// GetLastRequestUUID 根据 host_id 和 container_id 获取最近一次的 request uuid
func (s *SysdigLog) GetLastRequestUUID() string {
	if s.ProcessName == "fwatchdog" {
		if value, exist := conf.OfwatchdogRequestUUIDMap[s.HostID+"#"+s.ContainerID]; exist {
			if len(value) == 0 {
				return UNKNOWN
			}
			return helper.JoinKeys(value, ",")
		}
		return UNKNOWN
	}
	if value, exist := conf.NodeLastRequestUUIDMap[s.HostID+"#"+s.ContainerID]; exist {
		if value == "" { // 当前uuid被清空
			return UNKNOWN
		}
		return value
	}
	return UNKNOWN
}
