package service

import (
	"erinyes/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
)

type Data struct {
	ProcessCount   int      `json:"processCount"`
	FileCount      int      `json:"fileCount"`
	SocketCount    int      `json:"socketCount"`
	TotalNode      int      `json:"totalNode"`
	HostCount      int      `json:"hostCount"`
	ContainerCount int      `json:"containerCount"`
	NetCount       int      `json:"netCount"`
	SysdigCount    int      `json:"sysdigCount"`
	TotalEdge      int      `json:"totalEdge"`
	Top5UUID       []string `json:"top5UUID"`
	Top5Count      []int    `json:"top5Count"`
	Top10Syscall   []string `json:"top10Syscall"`
	Top10SysCount  []int    `json:"top10SysCount"`
}

// HandleDashboard 返回数据库中的主机数量、容器数量、顶点数量（进程、文件和套接字数量）和边（流量日志、审计日志）数量、产生活动最多的5个请求
func HandleDashboard(c *gin.Context) {
	var data Data
	hostSet := make(map[string]int)
	containerSet := make(map[string]int)
	db := models.GetMysqlDB()
	pageSize := 500 // 分页查，防止内存消耗太大

	pageNumber := 1
	for {
		var processes []models.Process
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&processes)
		if len(processes) == 0 {
			break
		}
		data.ProcessCount += len(processes)
		for _, process := range processes { // 遍历所有进程顶点
			hostSet[process.HostID] += 1
			containerSet[process.ContainerID] += 1
		}
		pageNumber++
	}

	pageNumber = 1
	for {
		var files []models.File
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&files)
		if len(files) == 0 {
			break
		}
		data.FileCount += len(files)
		for _, file := range files { // 遍历所有进程顶点
			hostSet[file.HostID] += 1
			containerSet[file.ContainerID] += 1
		}
		pageNumber++
	}

	pageNumber = 1
	for {
		var sockets []models.Socket
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&sockets)
		if len(sockets) == 0 {
			break
		}
		data.SocketCount += len(sockets)
		for _, socket := range sockets { // 遍历所有进程顶点
			hostSet[socket.HostID] += 1
			containerSet[socket.ContainerID] += 1
		}
		pageNumber++
	}
	data.TotalNode = data.ProcessCount + data.FileCount + data.SocketCount
	data.HostCount = len(hostSet)
	data.ContainerCount = len(containerSet)

	pageNumber = 1
	syscallMap := make(map[string]int)
	uuidMap := make(map[string]int)
	for {
		var events []models.Event
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&events)
		if len(events) == 0 {
			break
		}
		data.SysdigCount += len(events)
		for _, event := range events {
			if event.UUID != "" && event.UUID != "unknown" {
				uuidMap[event.UUID] += 1
			}
			syscallMap[event.Relation] += 1
		}
		pageNumber++
	}

	pageNumber = 1
	for {
		var nets []models.Net
		db.Order("id").Limit(pageSize).Offset((pageNumber - 1) * pageSize).Find(&nets)
		if len(nets) == 0 {
			break
		}
		data.NetCount += len(nets)
		for _, net := range nets {
			if net.UUID != "" && net.UUID != "unknown" {
				uuidMap[net.UUID] += 1
			}
		}
		pageNumber++
	}

	data.TotalEdge = data.SysdigCount + data.NetCount

	uuidSlice := SortMap(uuidMap)
	if len(uuidSlice) <= 5 {
		data.Top5UUID = uuidSlice
	} else {
		data.Top5UUID = uuidSlice[:5]
	}
	for _, uuid := range data.Top5UUID {
		data.Top5Count = append(data.Top5Count, uuidMap[uuid])
		//fmt.Println("%s: %d", uuid, uuidMap[uuid])
	}

	syscallSlice := SortMap(syscallMap)
	if len(syscallSlice) <= 10 {
		data.Top10Syscall = syscallSlice
	} else {
		data.Top10Syscall = syscallSlice[:10]
	}
	for _, sys := range data.Top10Syscall {
		data.Top10SysCount = append(data.Top10SysCount, syscallMap[sys])
	}

	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": data})
	return
}

func SortMap(uuidMap map[string]int) []string {
	var uuidSlice []string
	for key := range uuidMap {
		uuidSlice = append(uuidSlice, key)
	}

	sort.Slice(uuidSlice, func(i, j int) bool {
		return uuidMap[uuidSlice[i]] > uuidMap[uuidSlice[j]]
	})

	return uuidSlice
}
