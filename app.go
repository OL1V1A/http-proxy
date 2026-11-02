package main

import (
	"context"
	"fmt"
	"os"

	"http-proxy/config"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppConfig 是 UI 与后端之间传递的配置结构
type AppConfig struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	ProxyAddr string `json:"proxyAddr"`
}

// App 是暴露给前端的 Wails 绑定对象
type App struct {
	ctx context.Context
	mgr *ProxyManager
}

func NewApp() *App {
	return &App{mgr: NewProxyManager()}
}

// startup 在应用启动时由 Wails 调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetConfig 读取已保存的配置，文件不存在或使用默认值
func (a *App) GetConfig() AppConfig {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return AppConfig{Host: "0.0.0.0", Port: "3100", ProxyAddr: "http://127.0.0.1:11434"}
	}
	proxyAddr := cfg.Proxy.Address
	if proxyAddr == "" {
		proxyAddr = "http://127.0.0.1:11434"
	}
	port := cfg.Server.Port
	if port == "" {
		port = "3100"
	}
	return AppConfig{Host: cfg.Server.Host, Port: port, ProxyAddr: proxyAddr}
}

// IsRunning 返回代理服务是否在运行
func (a *App) IsRunning() bool {
	return a.mgr.IsRunning()
}

// Start 保存配置并启动代理服务，返回监听地址
func (a *App) Start(req AppConfig) (string, error) {
	if req.Port == "" {
		return "", fmt.Errorf("监听端口不能为空")
	}
	host := req.Host
	if host == "" {
		host = "0.0.0.0"
	}
	if req.ProxyAddr == "" {
		return "", fmt.Errorf("转发目标地址不能为空")
	}

	if err := saveConfigFile("config.yaml", req); err != nil {
		runtime.LogErrorf(a.ctx, "保存配置失败: %v", err)
	}

	logf := func(level, msg string) {
		runtime.EventsEmit(a.ctx, "log", map[string]string{"level": level, "msg": msg})
		if level == "error" {
			runtime.LogError(a.ctx, msg)
		} else {
			runtime.LogInfo(a.ctx, msg)
		}
	}

	addr := fmt.Sprintf("%s:%s", host, req.Port)
	if err := a.mgr.Start(host, req.Port, req.ProxyAddr, logf); err != nil {
		return "", err
	}
	runtime.EventsEmit(a.ctx, "status", true)
	return addr, nil
}

// Stop 停止代理服务
func (a *App) Stop() error {
	err := a.mgr.Stop()
	runtime.EventsEmit(a.ctx, "status", false)
	return err
}

// saveConfigFile 把当前配置写回 config.yaml
func saveConfigFile(path string, cfg AppConfig) error {
	host := cfg.Host
	if host == "" {
		host = "0.0.0.0"
	}
	data := fmt.Sprintf("server:\n  host: %s\n  port: %s\n\nproxy:\n  addr: %s\n", host, cfg.Port, cfg.ProxyAddr)
	return os.WriteFile(path, []byte(data), 0644)
}
