package service

import (
	"erinyes/parser"
	"github.com/gin-gonic/gin"
	"net/http"
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


