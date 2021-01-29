package main

import (
  "fmt"
  "os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	slog "github.com/sirupsen/logrus"
)

func initLogging(logFile string) {
  var file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    fmt.Println("Could Not Open Log File : " + err.Error())
  }
  slog.SetOutput(file)

  slog.SetFormatter(&slog.TextFormatter{})
}

func main() {

  initLogging("discord.log")

	// Start listening to requests send from Grafana. This call is blocking so
	// it wont finish until Grafana shutsdown the process or the plugin choose
	// to exit close down by itself
	err := datasource.Serve(newDiscordDataSource())

	// Log any error if we could start the plugin.
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
