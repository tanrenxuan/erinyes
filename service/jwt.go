package service

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const SECRET = "secret-key-random-string"

func generateToken(username string) string {
	// 创建一个Token对象
	token := jwt.New(jwt.SigningMethodHS256)

	// 设置token的自定义声明
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, _ := token.SignedString([]byte(SECRET))

	return tokenString
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	// 解析token字符串
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token的签名方法是否有效
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("无效的签名方法：%v", token.Header["alg"])
	}
	// 返回token中的声明部分
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("无效的Token")
}
