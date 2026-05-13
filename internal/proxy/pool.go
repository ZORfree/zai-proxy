package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"zai-proxy/internal/logger"
)

// proxyMode 表示代理获取方式
type proxyMode int

const (
	modeNone proxyMode = iota // 无代理
	modeFile                  // 从文件加载
	modeURL                   // 从远程 URL 实时获取
)

var (
	proxies     []string
	mu          sync.RWMutex
	currentMode proxyMode
	proxyURL    string
)

// InitProxies 初始化代理：优先使用文件，否则使用 URL 模式
func InitProxies(filePath string, url string) {
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，使用文件模式
		loadProxiesFromFile(filePath)
		currentMode = modeFile
		logger.LogInfo("Proxy mode: FILE (%s)", filePath)
	} else if url != "" {
		// 文件不存在，使用 URL 模式
		proxyURL = url
		currentMode = modeURL
		logger.LogInfo("Proxy mode: URL (%s)", url)
	} else {
		currentMode = modeNone
		logger.LogInfo("Proxy mode: NONE (no proxy file and no PROXY_URL)")
	}
}

// loadProxiesFromFile 从指定路径加载代理列表
// 格式: ip:port:username:password 或 ip:port
func loadProxiesFromFile(path string) {
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

// proxyAPIResponse 远程代理 API 返回的 JSON 结构
type proxyAPIResponse struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// fetchProxyFromURL 从远程 URL 获取一个代理，超时 2 秒
func fetchProxyFromURL() *proxyInfo {
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(proxyURL)
	if err != nil {
		logger.LogWarn("Failed to fetch proxy from URL: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LogWarn("Proxy URL returned status %d", resp.StatusCode)
		return nil
	}

	var apiResp proxyAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		logger.LogWarn("Failed to parse proxy URL response: %v", err)
		return nil
	}

	if apiResp.IP == "" || apiResp.Port == 0 {
		logger.LogWarn("Proxy URL response missing ip or port")
		return nil
	}

	addr := fmt.Sprintf("%s:%d", apiResp.IP, apiResp.Port)
	logger.LogDebug("Fetched proxy from URL: %s", addr)
	return &proxyInfo{addr: addr}
}

// getRandomProxyFromFile 从文件加载的列表中随机返回一个代理信息
func getRandomProxyFromFile() *proxyInfo {
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

// getProxyInfo 根据当前模式获取一个代理
func getProxyInfo() *proxyInfo {
	switch currentMode {
	case modeFile:
		return getRandomProxyFromFile()
	case modeURL:
		return fetchProxyFromURL()
	default:
		return nil
	}
}

// GetHTTPClient 返回一个配置了随机 SOCKS5 代理的 http.Client
func GetHTTPClient() *http.Client {
	info := getProxyInfo()
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
