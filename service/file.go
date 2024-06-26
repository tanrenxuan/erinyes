package service

import (
	"erinyes/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QueryFile struct {
	CurrPage int    `json:"curr_page"`
	PageSize int    `json:"page_size"`
	Query    string `json:"query"`
}

type DataFile struct {
	Files []models.File `json:"files"`
	Total int64         `json:"total"`
}

func HandleFile(c *gin.Context) {
	var req QueryFile
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
	db.Model(&models.File{}).Where("file_path LIKE ?", query).Count(&total)

	var files []models.File
	offset := (page - 1) * pageSize
	db.Where("file_path LIKE ?", query).Order("id").Offset(offset).Limit(pageSize).Find(&files)
	var data DataFile
	data.Files = files
	data.Total = total
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": data})
	return
}
