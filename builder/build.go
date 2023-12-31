package builder

import (
	"erinyes/logs"
	"erinyes/models"
	"erinyes/parser"
	"github.com/awalterschulze/gographviz"
	"os"
)

func createDir(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			logs.Logger.Errorf(err.Error())
		}
	}
}

// GenerateDotGraph 生成内存中的dot
func GenerateDotGraph() *gographviz.Graph {
	graphAst, _ := gographviz.Parse([]byte(`digraph G{}`))
	graph := gographviz.NewGraph()
	gographviz.Analyse(graphAst, graph)
	db := models.GetMysqlDB()

	// 遍历 Event 表和 Net 表
	pageSize := 100
	pageNumber := 1
	// 1. 遍历 Event 表
	for {
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
			GenerateEdge(start, end, event, graph)
		}
		pageNumber++
	}
	// 2. 遍历 Net 表
	pageNumber = 1
	for {
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
			GenerateEdge(startSocket, endSocket, net, graph)
		}
		pageNumber++
	}
	return graph
}

// GenerateDot 生成dot图文件
func GenerateDot(fileName string) {
	createDir("graphs/")
	dotName := "graphs/" + fileName + ".dot"
	graph := GenerateDotGraph()
	// 写入文件中
	fo, err := os.OpenFile(dotName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		logs.Logger.WithError(err).Fatalf("Open file %s failed", dotName)
	}
	defer fo.Close()
	fo.WriteString(graph.String())
}

// GenerateEdge 在图中生成一条边
func GenerateEdge(startVertex models.DotVertex, endVertex models.DotVertex, edge models.DotEdge, graph *gographviz.Graph) {
	// 基于 UUID 过滤一部分边
	if !edge.HasEdgeUUID() {
		return
	}

	// 边属性
	edgeM := make(map[string]string)
	edgeM["label"] = edge.EdgeName()
	// 起点属性
	startVertexM := make(map[string]string)
	startVertexM["shape"] = startVertex.VertexShape()
	// 重点属性
	endVertexM := make(map[string]string)
	endVertexM["shape"] = endVertex.VertexShape()

	// 加入起点所在子图
	startSubG := startVertex.VertexClusterID()
	startSubGM := make(map[string]string)
	startSubGM["label"] = startVertex.VertexClusterID()
	graph.AddSubGraph("G", startSubG, startSubGM)
	// 将起点加入起点子图
	graph.AddNode(startSubG, startVertex.VertexName(), startVertexM)

	// 加入终点所在子图
	endSubG := endVertex.VertexClusterID()
	endSubGM := make(map[string]string)
	endSubGM["label"] = endVertex.VertexClusterID()
	graph.AddSubGraph("G", endSubG, endSubGM)
	// 将终点加入终点子图
	graph.AddNode(endSubG, endVertex.VertexName(), endVertexM)

	// 加入边
	graph.AddEdge(startVertex.VertexName(), endVertex.VertexName(), true, edgeM)
}
