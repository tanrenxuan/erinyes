package logs

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger *logrus.Logger

func Init() {
	Logger = logrus.New()
	file, err := os.OpenFile("logs/mylogs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err.Error())
	}
	Logger.SetOutput(io.MultiWriter(file, os.Stdout))
}
