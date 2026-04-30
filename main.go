package main

import (
	"fmt"
	"net/http"

	"zai-proxy/internal/config"
	"zai-proxy/internal/handler"
	"zai-proxy/internal/logger"
	"zai-proxy/internal/proxy"
	"zai-proxy/internal/version"
)

var Version = "dev"

func main() {
	config.LoadConfig()
	logger.InitLogger()
	proxy.LoadProxies("data/proxies.txt")
	version.StartVersionUpdater()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"running", "version":"%s", "info":"Directory volume & TLS fix applied", "upstream_fe_version":"%s"}`+"\n", Version, version.GetFeVersion())
	})

	http.HandleFunc("/v1/models", handler.HandleModels)
	http.HandleFunc("/v1/chat/completions", handler.HandleChatCompletions)
	http.HandleFunc("/v1/messages", handler.HandleMessages)

	addr := ":" + config.Cfg.Port
	logger.LogInfo("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.LogError("Server failed: %v", err)
	}
}
