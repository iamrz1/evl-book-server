package cmd

import (
	"context"
	"evl-book-server/auth"
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
	"github.com/urfave/negroni"
)

func init() {
	serveCmd.PersistentFlags().IntP("port", "p", config.App().Port, "port on which the server will listen for http")

	err := viper.BindPFlag("app.port", serveCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		logger.Panicln("error binding flag", err)
	}
	rootCmd.AddCommand(serveCmd)
}

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

// serves the server
func serve(_ *cobra.Command, _ []string) {
	if !db.IsRedisUp() {
		logger.Println("redis server is down")
	}

	var router = mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/").HandlerFunc(routes.HomePageHandler)

	api := router.PathPrefix("/api").Subrouter().StrictSlash(true)
	api.HandleFunc("/login", routes.LoginHandler)
	api.HandleFunc("/signup", routes.AddUserHandler)
	api.HandleFunc("/validate/username/{username}", routes.ValidateUser)

	// api endpoints that goes through user auth middleware.
	// the user auth middleware uses token to authorize requests
	userAuthMW := negroni.New()
	userAuthMW.Use(&auth.Auth{})
	api.Handle("/update-profile", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.UpdateInfoHandler))))
	api.Handle("/books", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllBooksHandler))))
	api.Handle("/book/{id}", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetBookHandler))))
	api.Handle("/authors", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllAuthorsHandler))))
	api.Handle("/author/{id}", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAuthorHandler))))
	api.HandleFunc("/test", routes.ImageUploadHandOffHandler)
	api.Handle("/upload", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.ImageUploadHandOffHandler))))
	api.Handle("/upload/finalize", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.ImageUploadHandler))))

	api.Handle("/loan/request/{book_id}", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.CreateLoanRequestHandler))))
	api.Handle("/loans", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllLoansForThisUserHandler))))
	api.Handle("/loan/{id}", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetLoanByIDForThisUserHandler))))
	api.Handle("/loans/pending", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllPendingLoansForThisUserHandler))))
	api.Handle("/loans/active", userAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllActiveLoansForThisUserHandler))))

	// api endpoints that goes through admin auth middleware.
	// admin auth middleware uses token to authorize requests
	adminAuthMW := negroni.New()
	adminAuthMW.Use(&auth.Admin{})
	adminApi := router.PathPrefix("/api/admin").Subrouter().StrictSlash(true)
	adminApi.Handle("/book/create", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.BookCreateHandler))))
	adminApi.Handle("/book/update", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.BookUpdateHandler))))
	adminApi.Handle("/book/delete/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.BookDeleteHandler))))

	adminApi.Handle("/author/create", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.AuthorCreateHandler))))
	adminApi.Handle("/author/update", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.AuthorUpdateHandler))))
	adminApi.Handle("/author/delete/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.AuthorDeleteHandler))))

	adminApi.Handle("/loans", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllLoansHandler))))
	adminApi.Handle("/loan/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetLoanByIDHandler))))
	adminApi.Handle("/loans/pending", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllPendingLoansHandler))))
	adminApi.Handle("/loans/active", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.GetAllActiveLoansHandler))))
	adminApi.Handle("/loans/approve/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.ApproveLoanRequestHandler))))
	adminApi.Handle("/loans/decline/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.DeclineLoanRequestHandler))))
	adminApi.Handle("/loans/returned/{id}", adminAuthMW.With(negroni.Wrap(http.HandlerFunc(routes.ReturnedBookHandler))))

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
