package proxy

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/proxy"

	"zai-proxy/internal/logger"
)

var (
	proxies []string
	mu      sync.RWMutex
)

// LoadProxies 从指定路径加载代理列表
// 格式: ip:port:username:password 或 ip:port
func LoadProxies(path string) {
	file, err := os.Open(path)
	if err != nil {
		logger.LogInfo("No proxy file found at %s, running without proxy", path)
		return
	}
	defer file.Close()

	var loaded []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		loaded = append(loaded, line)
	}

	mu.Lock()
	proxies = loaded
	mu.Unlock()

	logger.LogInfo("Loaded %d proxies from %s", len(loaded), path)
}

type proxyInfo struct {
	addr     string
	username string
	password string
}

// getRandomProxyInfo 随机返回一个代理信息
func getRandomProxyInfo() *proxyInfo {
	mu.RLock()
	defer mu.RUnlock()

	if len(proxies) == 0 {
		return nil
	}

	line := proxies[rand.Intn(len(proxies))]
	parts := strings.Split(line, ":")

	switch len(parts) {
	case 2:
		return &proxyInfo{addr: fmt.Sprintf("%s:%s", parts[0], parts[1])}
	case 4:
		return &proxyInfo{
			addr:     fmt.Sprintf("%s:%s", parts[0], parts[1]),
			username: parts[2],
			password: parts[3],
		}
	default:
		logger.LogWarn("Invalid proxy format: %s", line)
		return nil
	}
}

// GetHTTPClient 返回一个配置了随机 SOCKS5 代理的 http.Client
func GetHTTPClient() *http.Client {
	info := getRandomProxyInfo()
	if info == nil {
		return &http.Client{}
	}

	logger.LogDebug("Using SOCKS5 proxy: %s", info.addr)

	var auth *proxy.Auth
	if info.username != "" {
		auth = &proxy.Auth{
			User:     info.username,
			Password: info.password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", info.addr, auth, proxy.Direct)
	if err != nil {
		logger.LogWarn("Failed to create SOCKS5 dialer: %v", err)
		return &http.Client{}
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		logger.LogWarn("SOCKS5 dialer does not support ContextDialer")
		return &http.Client{}
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return contextDialer.DialContext(ctx, network, addr)
			},
		},
	}
}
