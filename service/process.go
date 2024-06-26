package service

import (
	"erinyes/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QueryProcess struct {
	CurrPage int    `json:"curr_page"`
	PageSize int    `json:"page_size"`
	Query    string `json:"query"`
}

type DataProcess struct {
	Processes []models.Process `json:"processes"`
	Total     int64            `json:"total"`
}

func HandleProcess(c *gin.Context) {
	var req QueryProcess
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40001, "message": err.Error()})
		return
	}

	// 分页参数处理
	page := req.CurrPage
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	// 构建模糊查询语句
	query := "%" + req.Query + "%"
	var total int64
	db := models.GetMysqlDB()
	db.Model(&models.Process{}).Where("process_name LIKE ?", query).Count(&total)

	var processes []models.Process
	offset := (page - 1) * pageSize
	db.Where("process_name LIKE ?", query).Order("id").Offset(offset).Limit(pageSize).Find(&processes)
	var data DataProcess
	data.Processes = processes
	data.Total = total
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": data})
	return
}
