package cmd

import (
	"context"
	"evl-book-server/db"
	"evl-book-server/routes"
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"evl-book-server/config"
	"github.com/gorilla/mux"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the paceg http server",
	Run:   serve,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check if the port is available
		appCfg := config.App()
		portStr := strconv.Itoa(appCfg.Port)
		listener, err := net.Listen("tcp", ":"+portStr)
		if err != nil {
			return fmt.Errorf("port %s is not available", portStr)
		}
		_ = listener.Close()

		return nil
	},
}

func init() {
	serveCmd.PersistentFlags().IntP("port", "p", config.App().Port, "port on which the server will listen for http")

	err := viper.BindPFlag("app.port", serveCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		logger.Panicln("error binding flag", err)
	}
	rootCmd.AddCommand(serveCmd)
}

// serves the server
func serve(cmd *cobra.Command, args []string) {
	if !db.IsRedisUp() {
		logger.Println("redis server is down")
	}

	var router = mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", routes.HomePageHandler)

	appCfg := config.App()

	server := &http.Server{
		ReadTimeout:  appCfg.ReadTimeout,
		WriteTimeout: appCfg.WriteTimeout,
		IdleTimeout:  appCfg.IdleTimeout,
		Addr:         fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler:      router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error(err)
			os.Exit(-1)
		}
	}()

	logger.Info("Listening on <host> port" + fmt.Sprintf(":%d", viper.GetInt("app.port")))
	<-stop

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)

	logger.Info("Server shutdowns gracefully")
}
