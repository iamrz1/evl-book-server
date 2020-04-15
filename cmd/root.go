package cmd

import (
	"evl-book-server/db"
	"github.com/spf13/cobra"
	"log"
)

var (
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "server is a http book server",
	}
)

// Execute executes the root command of the evl-book-server
func Execute() {
	db.InitRedis()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
	//serve()
}
