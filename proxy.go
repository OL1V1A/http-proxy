package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogFunc 用于把代理运行日志回传到调用方（如 UI）
type LogFunc func(level, msg string)

// ProxyManager 负责在后台启动/停止 gin 反向代理服务
type ProxyManager struct {
	mu      sync.Mutex
	server  *http.Server
	running bool
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{}
}

// IsRunning 返回当前是否在运行
func (m *ProxyManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// Start 启动反向代理；同一时间仅允许一个实例
func (m *ProxyManager) Start(host, port, proxyAddr string, logf LogFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return fmt.Errorf("服务已在运行中")
	}

	target, err := url.Parse(proxyAddr)
	if err != nil {
		return fmt.Errorf("转发地址无效: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(response *http.Response) error {
		logf("info", fmt.Sprintf("响应: %s -> %d", response.Request.URL.Path, response.StatusCode))
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		logf("error", fmt.Sprintf("代理错误: %v", e))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(e.Error()))
	}

	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "http-proxy server is ok"})
	})

	// 未匹配到的请求（除 /health 外）全部转发到目标地址
	r.NoRoute(gin.WrapF(proxy.ServeHTTP))

	addr := fmt.Sprintf("%s:%s", host, port)

	// 启动前先探测端口是否可用，给出即时明确的反馈
	if ln, err := net.Listen("tcp", addr); err != nil {
		return fmt.Errorf("端口 %s 不可用（可能已被占用）: %w", addr, err)
	} else {
		_ = ln.Close()
	}

	srv := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	m.server = srv
	m.running = true

	go func() {
		logf("info", fmt.Sprintf("代理服务启动: http://%s", addr))
		logf("info", fmt.Sprintf("转发目标: %s", proxyAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logf("error", fmt.Sprintf("服务异常: %v", err))
			m.mu.Lock()
			m.running = false
			m.server = nil
			m.mu.Unlock()
		}
	}()

	return nil
}

// Stop 优雅停止代理服务
func (m *ProxyManager) Stop() error {
	m.mu.Lock()
	srv := m.server
	running := m.running
	m.mu.Unlock()

	if !running || srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)

	m.mu.Lock()
	m.running = false
	m.server = nil
	m.mu.Unlock()
	return err
}
