package main

import (
	"erinyes/builder"
	"erinyes/conf"
	"erinyes/logs"
	"erinyes/models"
	"erinyes/parser"
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
	parser.FileLogParse(false, sysdigFilepath, netFilepath)
}

func StartHTTP(_ *cobra.Command, args []string) {

}

// GenerateDot 可视化溯源图
func GenerateDot(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("no filepath after graph\n")
		os.Exit(-1)
	}
	builder.GenerateDot(args[0])
}
