package config

import (
	"time"

	"github.com/spf13/viper"
)

// Application represents the application config info
type Application struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Version      string
	Debug        bool
	Env          string
	Key          string
}

var appCfg Application

// LoadApp populates the app config instance
func LoadApp() {
	appCfg = Application{
		Port:         viper.GetInt("app.port"),
		ReadTimeout:  viper.GetDuration("app.read_timeout") * time.Second,
		WriteTimeout: viper.GetDuration("app.write_timeout") * time.Second,
		IdleTimeout:  viper.GetDuration("app.idle_timeout") * time.Second,
		Version:      viper.GetString("app.version"),
		Debug:        viper.GetBool("app.debug"),
		Env:          viper.GetString("app.env"),
		Key:          viper.GetString("app.key"),
	}
}

// App returns the app config instance
func App() Application {
	return appCfg
}
