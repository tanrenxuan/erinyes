package main

import (
	"database/sql"
	"fmt"
	"github.com/awalterschulze/gographviz"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const dataSourceName = "root:123456@tcp(localhost:3306)/erinyes"

const dir = "0124"

const defaultValue = "null"

func GetPaths() []string {
	var dotPaths []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// 只处理非目录类型的文件
			filename := filepath.Base(path)
			dotPaths = append(dotPaths, dir+"/"+filename)
		}
		return nil
	})
	return dotPaths
}

func GetGraph(dotPath string) *gographviz.Graph {
	f, err := os.ReadFile(dotPath)
	if err != nil {
		panic(err)
	}
	graph, err := gographviz.Read(f)
	if err != nil {
		panic(err)
	}
	return graph
}

func GetContainerId(dotPath string) string {
	splitList := strings.Split(dotPath, "-")
	return splitList[len(splitList)-2]
}

func GetTypeAndValue(s string) (string, string) {
	splitList := strings.Split(s, "##")
	return splitList[0], splitList[1]
}

func GetProcessTableId(value string, containerId string) int {
	db, err := sql.Open("mysql", dataSourceName)
	defer db.Close()
	if err != nil {
		panic(err)
	}
	query := "SELECT id FROM `process` WHERE container_id = ? and process_vpid = ?;"
	var id int

	err = db.QueryRow(query, containerId, value).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO `process` (host_id, host_name, container_id, container_name, process_vpid, process_name, process_exe_path) VALUES (?, ?, ?, ?, ?, ?, ?);"
		_, err = db.Exec(insertQuery, "ServerID", "ServerName", containerId, defaultValue, value, defaultValue, defaultValue)
		if err != nil {
			panic(err)
		}
	}

	query = "SELECT id FROM `process` WHERE container_id = ? and process_vpid = ?;"
	err = db.QueryRow(query, containerId, value).Scan(&id)

	return id
}

func GetFileTableId(value string, containerId string) int {
	db, err := sql.Open("mysql", dataSourceName)
	defer db.Close()
	if err != nil {
		panic(err)
	}
	query := "SELECT id FROM `file` WHERE container_id = ? and file_path = ?;"
	var id int

	err = db.QueryRow(query, containerId, value).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO `file` (host_id, host_name, container_id, container_name, file_path) VALUES (?, ?, ?, ?, ?);"
		_, err = db.Exec(insertQuery, "ServerID", "ServerName", containerId, defaultValue, value)
		if err != nil {
			panic(err)
		}
	}

	query = "SELECT id FROM `file` WHERE container_id = ? and file_path = ?;"
	err = db.QueryRow(query, containerId, value).Scan(&id)

	return id
}

func GetSocketTableId(value string, containerId string) int {
	db, err := sql.Open("mysql", dataSourceName)
	defer db.Close()
	if err != nil {
		panic(err)
	}

	ip, port := strings.Split(value, ":")[0], strings.Split(value, ":")[1]

	if port == "53" || port == "8080" {
		containerId = "outer"
	}

	query := "SELECT id FROM `socket` WHERE container_id = ? and dst_ip = ? and dst_port = ?;"
	var id int

	err = db.QueryRow(query, containerId, ip, port).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		insertQuery := "INSERT INTO `socket` (host_id, host_name, container_id, container_name, dst_ip, dst_port) VALUES (?, ?, ?, ?, ?, ?);"
		_, err = db.Exec(insertQuery, "ServerID", "ServerName", containerId, defaultValue, ip, port)
		if err != nil {
			panic(err)
		}
	}

	query = "SELECT id FROM `socket` WHERE container_id = ? and dst_ip = ? and dst_port = ?;"
	err = db.QueryRow(query, containerId, ip, port).Scan(&id)

	return id
}

func GetEventId(startId int, endId int, eventClass string, time string) error {
	db, err := sql.Open("mysql", dataSourceName)
	defer db.Close()
	if err != nil {
		panic(err)
	}

	timeBigInt, _ := strconv.ParseInt(time, 10, 64)

	if eventClass == "Socket2Socket" {
		insertQuery := "INSERT INTO `net` (src_id, dst_id, method, time) VALUES (?, ?, ?, ?);"
		_, err = db.Exec(insertQuery, startId, endId, "post", timeBigInt)
		if err != nil {
			panic(err)
		}
	} else {
		insertQuery := "INSERT INTO `event` (src_id, dst_id, event_class, relation, operation, time, uuid) VALUES (?, ?, ?, ?, ?, ?, ?);"
		_, err = db.Exec(insertQuery, startId, endId, eventClass, defaultValue, defaultValue, timeBigInt, defaultValue)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func transform(start string, end string, containerId string, time string) error {
	startType, startValue := GetTypeAndValue(start)
	endType, endValue := GetTypeAndValue(end)

	if startType == "Container" || endType == "Container" {
		return nil
	}

	var startId int
	var endId int

	if startType == "Process" {
		startId = GetProcessTableId(startValue, containerId)
	} else if startType == "File" {
		startId = GetFileTableId(startValue, containerId)
	} else if startType == "NetPeer" {
		startId = GetSocketTableId(startValue, containerId)
	}

	if endType == "Process" {
		endId = GetProcessTableId(endValue, containerId)
	} else if endType == "File" {
		endId = GetFileTableId(endValue, containerId)
	} else if endType == "NetPeer" {
		endId = GetSocketTableId(endValue, containerId)
	}

	eventClass := "null"
	if startType == "Process" && endType == "Process" {
		eventClass = "Process"
	} else if startType == "Process" && endType == "File" {
		eventClass = "File_V1"
	} else if startType == "Process" && endType == "NetPeer" {
		eventClass = "Network_V1"
	} else if startType == "File" && endType == "Process" {
		eventClass = "File_V2"
	} else if startType == "NetPeer" && endType == "Process" {
		eventClass = "Network_V2"
	} else if startType == "NetPeer" && endType == "NetPeer" {
		eventClass = "Socket2Socket"
	}

	GetEventId(startId, endId, eventClass, time)

	return nil
}

func main() {

	dotPaths := GetPaths()

	for _, dotPath := range dotPaths {
		graph := GetGraph(dotPath)

		containerId := GetContainerId(dotPath)

		fmt.Println("containerId=", containerId)

		for _, edge := range graph.Edges.Edges {
			start := edge.Src
			end := edge.Dst

			//fmt.Println("start=", start, "end=", end)
			//fmt.Println(GetTypeAndValue(start))
			//fmt.Println(GetTypeAndValue(end))

			start = strings.Trim(start, `"`)
			end = strings.Trim(end, `"`)

			time := strings.Trim(edge.Attrs["label"], `"`)

			transform(start, end, containerId, time)

		}
	}

}
