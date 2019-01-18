package server

import (
	"github.com/astaxie/beego/logs"
	"encoding/json"
)

func convertLogLevel(level string) int {
	switch (level) {
		case "debug":
			return logs.LevelDebug
		case "warn":
			return logs.LevelWarn
		case "error":
			return logs.LevelError
		case "info":
			return logs.LevelInfo
	}
	return logs.LevelDebug
}
func InitLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = G_config.logPath
	config["level"] = convertLogLevel(G_config.logLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}
