package service

import (
	"erinyes/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QuerySocket struct {
	CurrPage int    `json:"curr_page"`
	PageSize int    `json:"page_size"`
	Query    string `json:"query"`
}

type DataSocket struct {
	Sockets []models.Socket `json:"sockets"`
	Total   int64           `json:"total"`
}

func HandleSocket(c *gin.Context) {
	var req QuerySocket
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
	db.Model(&models.Socket{}).Where("dst_ip LIKE ?", query).Count(&total)

	var sockets []models.Socket
	offset := (page - 1) * pageSize
	db.Where("dst_ip LIKE ?", query).Order("id").Offset(offset).Limit(pageSize).Find(&sockets)
	var data DataSocket
	data.Sockets = sockets
	data.Total = total
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": data})
	return
}
