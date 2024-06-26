package service

import (
	"erinyes/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"Password"`
}

type Token struct {
	T string `json:"token"`
}

func HandleLogin(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 40001, "message": err.Error()})
		return
	}
	var user models.User
	db := models.GetMysqlDB()
	result := user.FindByName(db, req.Username)
	if result == false {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "用户不存在"})
		return
	}
	if user.Password != req.Password {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "密码错误"})
		return
	}
	token := generateToken(user.Username)
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success", "data": Token{T: token}})
	return
}

type Info struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func HandleInfo(c *gin.Context) {
	token := c.Query("token")
	fmt.Println(token)
	claims, err := parseToken(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 50014, "message": "token已过期，请重新登录"})
		return
	}
	username := claims["username"].(string)
	var user models.User
	db := models.GetMysqlDB()
	result := user.FindByName(db, username)
	if result == false {
		c.JSON(http.StatusOK, gin.H{"code": 40002, "message": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success",
		"data": Info{Name: username, Avatar: "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"}})
	return
}

func HandleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 20000, "message": "success"})
	return
}
