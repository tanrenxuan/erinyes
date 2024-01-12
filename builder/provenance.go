package builder

import (
	"erinyes/helper"
	"erinyes/logs"
	"erinyes/models"
	"erinyes/parser"
	"fmt"
	"gonum.org/v1/gonum/graph/multi"
	"strconv"
	"time"
)

const (
	ProcessTable = "process"
	FileTable    = "file"
	SocketTable  = "socket"
)

// RecordLoc 用来标识数据库中的一个顶点
type RecordLoc struct {
	Key   int    // primary key
	Table string // identify which table
}

// Provenance 根据 processID 溯源
func Provenance(hostID string, containerID string, processID string, processName string, timestamp *int, depth *int) *multi.WeightedDirectedGraph {
	// get root process
	mysqlDB := models.GetMysqlDB()
	var process models.Process
	if err := mysqlDB.First(&process, models.Process{HostID: hostID, ContainerID: containerID, ProcessVPID: processID, ProcessName: processName}).Error; err != nil {
		logs.Logger.WithError(err).Errorf("failed to build subgraph for process[host: %s, container: %s,process_vid: %s, process_name: %s]", hostID, containerID, processID, processName)
		return nil
	}
	g := multi.NewWeightedDirectedGraph()
	addedLine := make(map[int]bool)        // key为db/event中的primary id
	addedNode := make(map[RecordLoc]int64) // key为RecordLoc value为GraphNode的id
	id := AddNewGraphNode(g, Process,
		ProcessInfo{
			Path:          process.ProcessExepath,
			Name:          process.ProcessName,
			Pid:           process.ProcessVPID,
			ContainerID:   process.ContainerID,
			ContainerName: process.ContainerName,
			HostID:        process.HostID,
			HostName:      process.HostName})

	root := RecordLoc{Key: process.ID, Table: ProcessTable}
	addedNode[root] = id // 存入已访问顶点集合
	logs.Logger.Infof("开始构建溯源图，processID: %s, processName: %s, 其所在表: %s, 对应主键: %d", processID, process.ProcessName, ProcessTable, process.ID)
	logs.Logger.Infof("开始正向BFS溯源...")
	startTime := time.Now()
	BFS(g, root, addedLine, addedNode, false, depth)
	logs.Logger.Infof("It takes about %v seconds to forward BFS", time.Since(startTime).Seconds())
	middleTime := time.Now()
	logs.Logger.Infof("开始逆向BFS溯源...")
	BFS(g, root, addedLine, addedNode, true, depth)
	logs.Logger.Infof("It takes about %v seconds to backward BFS", time.Since(middleTime).Seconds())
	logs.Logger.Infof("子图构建成功...")
	logs.Logger.Infof("It takes about %v seconds to build Provenance Graph", time.Since(startTime).Seconds())
	return g
}

func AddNewGraphNode(g *multi.WeightedDirectedGraph, nodeType NodeType, nodeInfo NodeInfo) int64 {
	temp := g.NewNode()
	graphNode := GraphNode{
		id:       temp.ID(),
		nodeType: nodeType,
		nodeInfo: nodeInfo,
	}
	g.AddNode(graphNode)
	return temp.ID()
}

func AddNewGraphEdge(g *multi.WeightedDirectedGraph, from int64, to int64, relation string, timestamp int64, weight float64) {
	weightedLine := g.NewWeightedLine(GraphNode{id: from}, GraphNode{id: to}, weight)
	graphLine := GraphLine{
		F:         g.Node(from),
		T:         g.Node(to),
		W:         weightedLine.Weight(),
		Relation:  relation,
		TimeStamp: timestamp,
		UID:       weightedLine.ID(),
	}
	g.SetWeightedLine(graphLine)
}

// BFS 对数据库进行遍历，获取某个实体int的所有前向(后向)遍历子图(不包括root)
func BFS(g *multi.WeightedDirectedGraph, root RecordLoc, addedLine map[int]bool, addedNode map[RecordLoc]int64, reverse bool, maxLevel *int) {
	// 无需处理root
	visitedNode := map[RecordLoc]bool{root: true}
	var queue []RecordLoc
	currLevel := 0
	queue = append(queue, root)
	for {
		if len(queue) == 0 { // isEmpty(queue)
			break
		}
		if maxLevel != nil {
			if currLevel >= *maxLevel { // 到达指定遍历层数
				break
			}
		}
		size := len(queue)
		for i := 0; i < size; i++ { // 遍历当前层所有顶点（已经处理过）
			cur := queue[0]                                    // 必须用0 不能用i
			events := FetchEvents(cur.Key, cur.Table, reverse) // 寻找该顶点出发的所有边
			for _, e := range events {
				// 先存顶点
				var tempRecord RecordLoc
				tableName, err := GetTableName(e.EventClass, reverse)
				if err != nil {
					logs.Logger.WithError(err).Errorf("failed to get table name")
					continue // 忽略这条边以及对应的顶点
				}
				if reverse { // 逆向
					tempRecord = RecordLoc{Key: e.SrcID, Table: tableName}
				} else { // 正向
					tempRecord = RecordLoc{Key: e.DstID, Table: tableName}
				}
				if _, ok := visitedNode[tempRecord]; !ok { // 该顶点没有访问过
					// 判断该顶点是否已经存在于图中
					if _, ok := addedNode[tempRecord]; ok { // 该顶点已经在图中
						visitedNode[tempRecord] = true
						queue = append(queue, tempRecord) // 该顶点已经存在于图中，但依然需要遍历一次（正向和逆向都经过该点，但后续路劲存在差异）
					} else { // 该顶点不在图中
						if nodeType, nodeInfo, err := GetEntityNode(tempRecord); err != nil {
							logs.Logger.WithError(err).Errorf("failed to fetch entity")
							continue // 不再考虑边
						} else {
							id := AddNewGraphNode(g, nodeType, nodeInfo) // 处理该顶点，加入图中
							visitedNode[tempRecord] = true
							addedNode[tempRecord] = id
							queue = append(queue, tempRecord) // 只有将该顶点成功加入Graph中，才将该顶点送入queue
						}
					}
				}
				// 再存边
				if _, ok := addedLine[e.ID]; !ok { // 该事件(line)没有访问过
					addedLine[e.ID] = true
					var (
						fromID int64
						toID   int64
						tempA  int64
						tempB  int64
					)
					if tempA, ok = addedNode[tempRecord]; !ok {
						panic("No such Node in the graph")
					}
					if tempB, ok = addedNode[cur]; !ok {
						panic("No such Node in the graph")
					}
					if reverse { // 逆向
						fromID = tempA
						toID = tempB
					} else { // 正向
						fromID = tempB
						toID = tempA
					}
					AddNewGraphEdge(g, fromID, toID, e.Relation, e.Time, 0) // weight暂时为空
				}
			}
			queue = queue[1:] // 删除头部
		}
		currLevel++
	}
}

func FetchEvents(key int, table string, reverse bool) []models.Event {
	mysqlDB := models.GetMysqlDB()
	// 根据该实体所在表推断其sql
	sqlStr := "event_class = ?"
	switch table {
	case ProcessTable:
		if reverse { // 1. process -> process 2. file -> process 3. socket -> process
			mysqlDB = mysqlDB.Where(sqlStr+" or "+sqlStr+" or "+sqlStr,
				parser.PROCESS, parser.FILEV2, parser.NETWORKV2)
		} else { // 1. process -> process 2. process -> file 3. process -> socket
			mysqlDB = mysqlDB.Where(sqlStr+" or "+sqlStr+" or "+sqlStr,
				parser.PROCESS, parser.FILEV1, parser.NETWORKV1)
		}
	case FileTable:
		if reverse { // 1. process -> file
			mysqlDB = mysqlDB.Where(sqlStr, parser.FILEV1)
		} else { // 1. file -> process
			mysqlDB = mysqlDB.Where(sqlStr, parser.FILEV2)
		}
	case SocketTable:
		if reverse { // 1. process -> socket
			mysqlDB = mysqlDB.Where(sqlStr, parser.NETWORKV1)
		} else { // 1. socket -> process
			mysqlDB = mysqlDB.Where(sqlStr, parser.NETWORKV2)
		}
	default:
		logs.Logger.Errorf("failed to parse table %s, fetch events failed", table)
		return nil
	}
	if reverse { // 逆向
		mysqlDB = mysqlDB.Where("dst_id = ?", strconv.Itoa(key))
	} else { // 正向
		mysqlDB = mysqlDB.Where("src_id = ?", strconv.Itoa(key))
	}
	var events []models.Event
	if err := mysqlDB.Find(&events).Error; err != nil {
		logs.Logger.WithError(err).Errorf("failed to fetch events(edges) from db")
		return nil
	}
	return events
}

// GetTableName 根据事件的eventClass推断应该从哪个表获取顶点
func GetTableName(eventClass string, reverse bool) (string, error) {
	switch eventClass {
	case parser.PROCESS: // process -> process
		return ProcessTable, nil
	case parser.FILEV1: // process -> file
		return helper.MyStringIf(reverse, ProcessTable, FileTable), nil
	case parser.FILEV2: // file -> process
		return helper.MyStringIf(reverse, FileTable, ProcessTable), nil
	case parser.NETWORKV1: // process -> socket
		return helper.MyStringIf(reverse, ProcessTable, SocketTable), nil
	case parser.NETWORKV2: // socket -> process
		return helper.MyStringIf(reverse, SocketTable, ProcessTable), nil
	}
	return "", fmt.Errorf("failed to calculate the table by eventClass: %s", eventClass)
}

func GetEntityNode(r RecordLoc) (NodeType, NodeInfo, error) {
	mysqlDB := models.GetMysqlDB()
	switch r.Table {
	case ProcessTable:
		var process models.Process
		if err := mysqlDB.First(&process, r.Key).Error; err != nil {
			return -1, nil, fmt.Errorf("failed to get process entity node from db, err: %w", err)
		}
		return Process, ProcessInfo{
			Path:          process.ProcessExepath,
			Name:          process.ProcessName,
			Pid:           process.ProcessVPID,
			HostName:      process.HostName,
			HostID:        process.HostID,
			ContainerName: process.ContainerName,
			ContainerID:   process.ContainerID}, nil
	case FileTable:
		var file models.File
		if err := mysqlDB.First(&file, r.Key).Error; err != nil {
			return -1, nil, fmt.Errorf("failed to get file entity node from db, err: %w", err)
		}
		return File, FileInfo{
			HostName:      file.HostName,
			HostID:        file.HostID,
			ContainerID:   file.ContainerID,
			ContainerName: file.ContainerName,
			Path:          file.FilePath}, nil
	case SocketTable:
		var socket models.Socket
		if err := mysqlDB.First(&socket, r.Key).Error; err != nil {
			return -1, nil, fmt.Errorf("failed to get socket entity node from db, err: %w", err)
		}
		return Socket, SocketInfo{
			DstIP:         socket.DstIP,
			DstPort:       socket.DstPort,
			ContainerName: socket.ContainerName,
			ContainerID:   socket.ContainerID,
			HostID:        socket.HostID,
			HostName:      socket.HostName}, nil
	}
	return -1, nil, fmt.Errorf("unknown record %s, can't find the entity from db", r.Table)
}
