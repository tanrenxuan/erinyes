package service

import (
	"archive/zip"
	"bytes"
	"erinyes/builder"
	"erinyes/logs"
	"erinyes/parser"
	"github.com/gin-gonic/gin"
	"net/http"
	"os/exec"
	"time"
)

func HandlePing(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to use graph build service")
}

type GeneralLogData struct {
	Log string `json:"log"`
}

type GeneralLogsData struct {
	Logs []string `json:"logs"`
}

func HandleSysdigLog(c *gin.Context) {
	var sysdigData GeneralLogData
	if err := c.ShouldBindJSON(&sysdigData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parser.SysdigRawChan <- sysdigData.Log
	c.String(http.StatusOK, "Add sysdig log to chan success")
}

func HandleSysdigLogs(c *gin.Context) {
	var sysdigData GeneralLogsData
	if err := c.ShouldBindJSON(&sysdigData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, value := range sysdigData.Logs {
		parser.SysdigRawChan <- value
	}

	c.String(http.StatusOK, "Add all sysdig logs to chan success")
}

func HandleNetLog(c *gin.Context) {
	var netData GeneralLogData
	if err := c.ShouldBindJSON(&netData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parser.NetRawChan <- netData.Log
	c.String(http.StatusOK, "Add net log to chan success")
}

func HandleNetLogs(c *gin.Context) {
	var netData GeneralLogsData
	if err := c.ShouldBindJSON(&netData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, value := range netData.Logs {
		parser.NetRawChan <- value
	}

	c.String(http.StatusOK, "Add all net logs to chan success")
}

func HandleGenerate(c *gin.Context) {
	currentTime := time.Now()
	currentTimeString := currentTime.Format("20060102150405")
	dotName := currentTimeString + ".dot"
	svgName := currentTimeString + ".svg"
	dotString := builder.GenerateDotGraph("").String()
	dotContent := []byte(dotString)
	err, svgContent := generateSVGFromDot(dotContent) // 替换为你生成 svg 文件的逻辑
	if err != nil {
		logs.Logger.WithError(err).Errorf("generate svg failed")
		c.String(http.StatusInternalServerError, "生成svg失败")
		return
	}

	// 创建一个内存中的 ZIP 文件
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// 添加 dot 文件到 ZIP 文件中
	dotFile, _ := zipWriter.Create(dotName)
	dotFile.Write(dotContent)

	// 添加 svg 文件到 ZIP 文件中
	svgFile, _ := zipWriter.Create(svgName)
	svgFile.Write(svgContent)

	// 关闭 ZIP 文件
	zipWriter.Close()

	// 设置响应头，指定文件名
	c.Header("Content-Disposition", "attachment; filename=files.zip")

	// 返回 ZIP 文件内容
	c.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}

func generateSVGFromDot(dotContent []byte) (error, []byte) {
	// 使用 exec.Command 执行 dot 命令生成 svg
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = bytes.NewReader(dotContent)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err, []byte{}
	}
	return nil, out.Bytes()
}
