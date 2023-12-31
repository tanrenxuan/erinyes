package conf

import (
	"erinyes/logs"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type ConfigStruct struct {
	Mysql struct {
		Host         string `yaml:"Host"`
		Port         int    `yaml:"Port"`
		Username     string `yaml:"Username"`
		Password     string `yaml:"Password"`
		DBName       string `yaml:"DBName"`
		MaxOpenConns int    `yaml:"MaxOpenConns"`
		MaxIdleConns int    `yaml:"MaxIdleConns"`
	} `yaml:"Mysql"`
	Service struct {
		Port string `yaml:"Port"`
	} `yaml:"Service"`
	IPMap      map[string]string `yaml:"IPMap"`
	GatewayMap map[string]bool   `yaml:"GatewayMap"`
}

var Config ConfigStruct

func Init() {
	yamlFile, err := ioutil.ReadFile("conf/config.yaml")
	if err != nil {
		logs.Logger.WithError(err).Fatal("read config file failed")
	}
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		logs.Logger.WithError(err).Fatal("unmarshal config file failed")
	}
	logs.Logger.Info("成功解析配置文件config")
	LastRequestUUIDMap = make(map[string]string)

}

const (
	MockHostID   = "ServerID"
	MockHostName = "ServerName"
	//OuterHostID        = "OuterHostID"
	//OuterHostName      = "OuterHostName"
	OuterContainerID   = "OuterContainerID"
	OuterContainerName = "OuterContainerName"
)

var LastRequestUUIDMap map[string]string // host_id#容器id -> lastRequestUUID
