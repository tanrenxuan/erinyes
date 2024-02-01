package main

import (
	"erinyes/builder"
	"erinyes/conf"
	"erinyes/logs"
	"erinyes/models"
	"erinyes/parser"
	"erinyes/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/graph/multi"
	"os"
	"strconv"
)

func main() {
	logs.Init()
	conf.Init()
	models.Init()

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand([]*cobra.Command{
		{
			Use:                "service",
			Short:              "Start a http service to get log and parse, build graph",
			DisableFlagParsing: true,
			Run:                StartHTTP,
		},
		{
			Use:                "graph",
			Short:              "Generate graph in db",
			DisableFlagParsing: true,
			Run:                GenerateGraph,
		},
		{
			Use:                "dot",
			Short:              "Generate dot file",
			DisableFlagParsing: true,
			Run:                GenerateDot,
		},
		{
			Use:                "subgraph",
			Short:              "Build sub provenance graph for certain process which identified by process id and host and container",
			DisableFlagParsing: true,
			Run:                BuildSubGraph,
		},
	}...)
	if err := rootCmd.Execute(); err != nil {
		logs.Logger.WithError(err).Fatal("failed to run command")
	}
}

func GenerateGraph(_ *cobra.Command, args []string) {
	var (
		sysdigFilepath string
		netFilepath    string
	)
	if len(args) == 0 {
		fmt.Printf("no filepath after graph\n")
		os.Exit(-1)
	} else if len(args) == 1 {
		sysdigFilepath = args[0]
		netFilepath = ""
	} else {
		sysdigFilepath = args[0]
		netFilepath = args[1]
	}
	parser.FileLogParse(true, sysdigFilepath, netFilepath)
}

func StartHTTP(_ *cobra.Command, args []string) {
	go parser.HTTPLogParse(true)
	r := gin.Default()

	r.GET("/api/ping", service.HandlePing)
	r.POST("/api/sysdig/log", service.HandleSysdigLog)
	r.POST("/api/sysdig/logs", service.HandleSysdigLogs)

	r.POST("/api/net/log", service.HandleNetLog)
	r.POST("/api/net/logs", service.HandleNetLogs)

	r.GET("/api/generate", service.HandleGenerate)
	err := r.Run(conf.Config.Service.Port)
	if err != nil {
		panic(err)
	}

}

// GenerateDot 可视化溯源图
func GenerateDot(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("no filepath after graph\n")
		os.Exit(-1)
	}
	if len(args) == 2 {
		builder.GenerateDot(args[0], args[1])
	} else {
		builder.GenerateDot(args[0], "")
	}
}

func BuildSubGraph(cmd *cobra.Command, args []string) {
	if !(len(args) == 5 || len(args) == 6) {
		fmt.Printf("construct cmd must need host, container and process id, depth optional.\n")
		logs.Logger.Errorf("construct graph failed, args = %s", args)
		return
	}
	var g *multi.WeightedDirectedGraph
	timeLimit := true
	uuid := "62c791cd9a1db361c0811cc5bfee2f02" // Erinyes high
	//uuid := ""
	if len(args) == 6 {
		depth, err := strconv.Atoi(args[5])
		if err == nil {
			g = builder.Provenance(args[0], args[1], args[2], args[3], nil, &depth, timeLimit, uuid)
		} else {
			fmt.Printf("depth is not valid, use default depth.\n")
			g = builder.Provenance(args[0], args[1], args[2], args[3], nil, nil, timeLimit, uuid)
		}
	} else {
		fmt.Printf("depth not absent, use default depth.\n")
		g = builder.Provenance(args[0], args[1], args[2], args[3], nil, nil, timeLimit, uuid)
	}
	if g == nil {
		logs.Logger.Infof("failed to get provenance graph")
		return
	}
	if err := builder.Visualize(g, args[4]); err != nil {
		logs.Logger.WithError(err).Errorf("failed to visualize provenance graph")
		fmt.Printf("Visualize provenance graph for %s failed, err = %s", args[3], err.Error())
		return
	}
	fmt.Printf("Visualize provenance graph for %s success!", args[2])
	logs.Logger.Infof("success to visualize provenance graph")
}
