package service

import (
	"erinyes/helper"
	"erinyes/logs"
	"erinyes/models"
	"erinyes/parser"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type QueryGraph struct {
	IfAllGraph  bool   `json:"ifAllGraph"` // 若为true，则返回全图；否则，根据指定进程节点进行查询
	UUID        string `json:"uuid"`       // 根据特定请求进行查询。若为空，则忽略。
	HostID      string `json:"hostID"`     // <HostID, ContainerID, VPid, ProcessName>唯一定位一个进程节点，只有IfAllGraph为false才有用
	ContainerID string `json:"containerID"`
	VPid        string `json:"vpid"`
	ProcessName string `json:"processName"`
}

type DataGraph struct { // 响应体
	Nodes      []Node         `json:"nodes"`
	Links      []Link         `json:"links"`
	Categories []Category     `json:"categories"`
	Stat       StatDetail     `json:"stat"`
	Syscalls   map[string]int `json:"syscalls"`
}

type StatDetail struct {
	HostNum      int `json:"host_num"`
	ContainerNum int `json:"container_num"`
	ProcessNum   int `json:"process_num"`
	FileNum      int `json:"file_num"`
	SocketNum    int `json:"socket_num"`
	EventNum     int `json:"event_num"`
	NetNum       int `json:"net_num"`
	SyscallNum   int `json:"syscall_num"`
}

type Node struct {
	ID       string `json:"id"`       // 顶点的唯一标识符
	Name     string `json:"name"`     // 顶点的Label
	Category int    `json:"category"` // 顶点的类别，值为Category数组的下标
	Symbol   string `json:"symbol"`   // 顶点的形状：rect、circle、diamond
	Info     string `json:"info"`     // 详情，用空行表示换行即可，前端会处理
}

type Link struct {
	Source string `json:"source"` // 起点，Node中的ID字段
	Target string `json:"target"` // 终点，Node中的ID字段
	Name   string `json:"name"`   //边的Label
	Info   string `json:"info"`   // 详情
}

type Category struct { // 类别按照容器进行区分
	Name string `json:"name"`
}

// HandleGraph 处理图数据的获取
func HandleGraph(c *gin.Context) {
	var req QueryGraph
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40001, "message": err.Error()})
		return
	}
	if req.IfAllGraph { // 搜索全图
		g := searchAllGraph(req.UUID, false)
		//fmt.Println(g)
		c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": g})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 40001, "message": "请设置IfAllGraph未true"})
	return
}

// searchAllGraph搜索全图
func searchAllGraph(uuid string, demo bool) DataGraph {
	db := models.GetMysqlDB()
	var graph DataGraph
	nodeMap := make(map[string]bool)    //顶点唯一标识符集合
	var nodeSlice []Node                // 存放所有的Node
	categoryMap := make(map[string]int) // 顶点的类别 -> 类别数组的下标
	var categorySlice []Category        // 存放所有的类别
	var linkSlice []Link

	processNum, fileNum, socketNum := 0, 0, 0
	syscallMap := make(map[string]int)

	// 遍历 Event 表和 Net 表
	pageSize := 100
	pageNumber := 1
	// 1. 遍历 Event 表
	for {
		if demo && pageNumber == 2 {
			break
		}
		var events []models.Event
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&events)
		if len(events) == 0 {
			break
		}
		for _, event := range events { // 遍历所有边
			var start models.DotVertex
			var end models.DotVertex
			switch event.EventClass { // eventType 决定了从哪两个表中查询数据
			case parser.PROCESS: // process -> process
				var startProcess models.Process
				var endProcess models.Process
				result1 := startProcess.FindByID(db, event.SrcID)
				result2 := endProcess.FindByID(db, event.DstID)
				if !(result1 && result2) {
					continue
				}
				start = startProcess
				end = endProcess
			case parser.FILEV1: // process -> file
				var startProcess models.Process
				var endFile models.File
				result1 := startProcess.FindByID(db, event.SrcID)
				result2 := endFile.FindByID(db, event.DstID)
				if !(result1 && result2) {
					continue
				}
				start = startProcess
				end = endFile
			case parser.FILEV2: // file -> process
				var startFile models.File
				var endProcess models.Process
				result1 := startFile.FindByID(db, event.SrcID)
				result2 := endProcess.FindByID(db, event.DstID)
				if !(result1 && result2) {
					continue
				}
				start = startFile
				end = endProcess
			case parser.NETWORKV1: // process -> socket
				var startProcess models.Process
				var endSocket models.Socket
				result1 := startProcess.FindByID(db, event.SrcID)
				result2 := endSocket.FindByID(db, event.DstID)
				if !(result1 && result2) {
					continue
				}
				start = startProcess
				end = endSocket
			case parser.NETWORKV2: // socket -> process
				var startSocket models.Socket
				var endProcess models.Process
				result1 := startSocket.FindByID(db, event.SrcID)
				result2 := endProcess.FindByID(db, event.DstID)
				if !(result1 && result2) {
					continue
				}
				start = startSocket
				end = endProcess
			default:
				logs.Logger.Warnf("Unknown event class: %s in event tables", event.EventClass)
			}
			r := generateLink(start, end, event, &linkSlice, uuid, &nodeMap, &nodeSlice, &categoryMap, &categorySlice, &processNum, &fileNum, &socketNum, &syscallMap)
			if r == true {
				graph.Stat.EventNum += 1
			}
		}
		pageNumber++
	}

	// 2. 遍历 Net 表
	pageNumber = 1
	for {
		if demo && pageNumber == 2 {
			break
		}
		var nets []models.Net
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&nets)
		if len(nets) == 0 {
			break
		}
		for _, net := range nets {
			var startSocket models.Socket
			var endSocket models.Socket
			result1 := startSocket.FindByID(db, net.SrcID)
			result2 := endSocket.FindByID(db, net.DstID)
			if !(result1 && result2) {
				continue
			}
			r := generateLink(startSocket, endSocket, net, &linkSlice, uuid, &nodeMap, &nodeSlice, &categoryMap, &categorySlice, &processNum, &fileNum, &socketNum, &syscallMap)
			if r == true {
				graph.Stat.NetNum += 1
			}
		}
		pageNumber++
	}
	graph.Links = linkSlice
	graph.Nodes = nodeSlice
	graph.Categories = categorySlice
	hostMap := make(map[string]bool)
	for _, value := range categorySlice {
		parts := strings.Split(value.Name, "_")
		hostMap[parts[0]] = true
	}
	graph.Stat.ContainerNum = len(categorySlice)
	graph.Stat.HostNum = len(hostMap)
	graph.Stat.ProcessNum = processNum
	graph.Stat.FileNum = fileNum
	graph.Stat.SocketNum = socketNum
	graph.Stat.SyscallNum = len(syscallMap)
	graph.Syscalls = syscallMap
	return graph
}

// generateLink 在结构体g中生成link
func generateLink(startVertex models.DotVertex, endVertex models.DotVertex, edge models.DotEdge,
	linkSlice *[]Link, uuid string, nodeMap *map[string]bool, nodeSlice *[]Node,
	categoryMap *map[string]int, categorySlice *[]Category, processNum *int, fileNum *int, socketNum *int, syscallMap *map[string]int) bool {
	if uuid != "" {
		if !edge.HasEdgeUUID() {
			return false
		}
		uuidStr := edge.GetUUID()
		uuids := strings.Split(uuidStr, ",")
		if !helper.SliceContainsTarget(uuids, uuid) {
			return false
		}
	}
	var l Link // 一定会产生一个连接，但不一定会有新的节点
	l.Name = edge.LinkLabel()
	l.Info = edge.LinkInfo()
	(*syscallMap)[edge.LinkLabel()] += 1
	if _, ok := (*nodeMap)[startVertex.LinkID()]; ok { // 起点已经存在，不需要创建
		l.Source = startVertex.LinkID()
	} else {
		var categoryIndex int
		if value, ok := (*categoryMap)[startVertex.LinkCategory()]; ok { // 类别存在于category数组中
			categoryIndex = value
		} else {
			*categorySlice = append(*categorySlice, Category{
				Name: startVertex.LinkCategory(),
			})
			(*categoryMap)[startVertex.LinkCategory()] = len(*categorySlice) - 1
			categoryIndex = len(*categorySlice) - 1
		}

		*nodeSlice = append(*nodeSlice, Node{
			ID:       startVertex.LinkID(),
			Name:     startVertex.LinkName(),
			Category: categoryIndex, // 应当写Category数组的下标
			Symbol:   startVertex.LinkSymbol(),
			Info:     startVertex.LinkInfo(),
		})
		if startVertex.LinkSymbol() == "rect" { // 进程
			*processNum += 1
		} else if startVertex.LinkSymbol() == "circle" { // 文件
			*fileNum += 1
		} else { // 套接字
			*socketNum += 1
		}
		(*nodeMap)[startVertex.LinkID()] = true // true没有意义，map当作set用
		l.Source = startVertex.LinkID()
	}

	if _, ok := (*nodeMap)[endVertex.LinkID()]; ok { // 终点已经存在，不需要创建
		l.Target = endVertex.LinkID()
	} else {
		var categoryIndex int
		if value, ok := (*categoryMap)[endVertex.LinkCategory()]; ok { // 类别存在于category数组中
			categoryIndex = value
		} else {
			*categorySlice = append(*categorySlice, Category{
				Name: endVertex.LinkCategory(),
			})
			(*categoryMap)[endVertex.LinkCategory()] = len(*categorySlice) - 1
			categoryIndex = len(*categorySlice) - 1
		}

		*nodeSlice = append(*nodeSlice, Node{
			ID:       endVertex.LinkID(),
			Name:     endVertex.LinkName(),
			Category: categoryIndex, // 应当写Category数组的下标
			Symbol:   endVertex.LinkSymbol(),
			Info:     endVertex.LinkInfo(),
		})
		if endVertex.LinkSymbol() == "rect" { // 进程
			*processNum += 1
		} else if endVertex.LinkSymbol() == "circle" { // 文件
			*fileNum += 1
		} else { // 套接字
			*socketNum += 1
		}
		(*nodeMap)[endVertex.LinkID()] = true // true没有意义，map当作set用
		l.Target = endVertex.LinkID()
	}
	*linkSlice = append(*linkSlice, l)
	return true
}
