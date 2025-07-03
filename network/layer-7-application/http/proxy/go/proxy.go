package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 代理服务器结构
type ProxyServer struct {
	targetURL *url.URL
	proxy     *httputil.ReverseProxy
}

// 创建新的代理服务器
func NewProxyServer(targetURL string) (*ProxyServer, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(url)

	// 自定义请求修改器（可选）
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 可以在这里修改响应头
		resp.Header.Set("X-Proxy-By", "Go-Proxy-Server")
		return nil
	}

	// 错误处理器
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("代理错误: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte("代理服务器错误"))
	}

	return &ProxyServer{
		targetURL: url,
		proxy:     proxy,
	}, nil
}

// 处理HTTP请求
func (p *ProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// 记录请求信息
	log.Printf("代理请求: %s %s -> %s", req.Method, req.URL.Path, p.targetURL.String())

	// 可以在这里添加请求预处理逻辑
	// 例如：添加认证头、修改请求路径等

	// 转发请求
	p.proxy.ServeHTTP(rw, req)
}

// 带有路径映射的代理服务器
type MultiProxyServer struct {
	routes map[string]*ProxyServer
}

// 创建多路由代理服务器
func NewMultiProxyServer() *MultiProxyServer {
	return &MultiProxyServer{
		routes: make(map[string]*ProxyServer),
	}
}

// 添加路由
func (m *MultiProxyServer) AddRoute(path string, targetURL string) error {
	proxy, err := NewProxyServer(targetURL)
	if err != nil {
		return err
	}
	m.routes[path] = proxy
	return nil
}

// 处理请求
func (m *MultiProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// 根据路径选择目标服务器
	for path, proxy := range m.routes {
		if len(req.URL.Path) >= len(path) && req.URL.Path[:len(path)] == path {
			// 移除路径前缀（可选）
			// req.URL.Path = req.URL.Path[len(path):]
			proxy.ServeHTTP(rw, req)
			return
		}
	}

	// 没有匹配的路由
	http.NotFound(rw, req)
}

// 目标服务器1 (端口8081)
func startTargetServer1(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器1的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Println("目标服务器1启动在端口8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("目标服务器1错误: %v", err)
		}
	}()

	// 等待关闭信号
	<-ctx.Done()
	log.Println("正在关闭目标服务器1...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("目标服务器1关闭错误: %v", err)
	} else {
		log.Println("目标服务器1已关闭")
	}
}

// 目标服务器2 (端口8082)
func startTargetServer2(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器2(静态资源)的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	go func() {
		log.Println("目标服务器2启动在端口8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("目标服务器2错误: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("正在关闭目标服务器2...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("目标服务器2关闭错误: %v", err)
	} else {
		log.Println("目标服务器2已关闭")
	}
}

// 目标服务器3 (端口8083)
func startTargetServer3(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器3(默认服务)的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	server := &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}

	go func() {
		log.Println("目标服务器3启动在端口8083")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("目标服务器3错误: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("正在关闭目标服务器3...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("目标服务器3关闭错误: %v", err)
	} else {
		log.Println("目标服务器3已关闭")
	}
}

// 启动代理服务器
func startProxyServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// 等待目标服务器启动
	time.Sleep(2 * time.Second)

	// 方式1: 单一目标代理
	fmt.Println("=== 单一目标代理示例 ===")
	singleProxy, err := NewProxyServer("http://localhost:8081")
	if err != nil {
		log.Fatal("创建代理失败:", err)
	}

	// 方式2: 多目标代理
	fmt.Println("=== 多目标代理示例 ===")
	multiProxy := NewMultiProxyServer()
	multiProxy.AddRoute("/api", "http://localhost:8081")
	multiProxy.AddRoute("/static", "http://localhost:8082")
	multiProxy.AddRoute("/", "http://localhost:8083") // 默认路由

	// 启动代理服务器
	mux := http.NewServeMux()

	// 单一代理路由
	mux.Handle("/single/", http.StripPrefix("/single", singleProxy))

	// 多目标代理路由
	mux.Handle("/", multiProxy)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		fmt.Println("代理服务器启动在端口 8080")
		fmt.Println("单一代理: http://localhost:8080/single/")
		fmt.Println("多目标代理:")
		fmt.Println("  /api -> http://localhost:8081")
		fmt.Println("  /static -> http://localhost:8082")
		fmt.Println("  / -> http://localhost:8083")
		fmt.Println("\n测试URL:")
		fmt.Println("  http://localhost:8080/api/test")
		fmt.Println("  http://localhost:8080/static/css/style.css")
		fmt.Println("  http://localhost:8080/home")
		fmt.Println("  http://localhost:8080/single/target1/test")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("代理服务器错误: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("正在关闭代理服务器...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("代理服务器关闭错误: %v", err)
	} else {
		log.Println("代理服务器已关闭")
	}
}

// 简单版本 - 同时启动所有服务器
func mainSimple() {
	var wg sync.WaitGroup

	// 启动目标服务器
	wg.Add(1)
	go startTargetServer1Simple(&wg)

	wg.Add(1)
	go startTargetServer2Simple(&wg)

	wg.Add(1)
	go startTargetServer3Simple(&wg)

	// 启动代理服务器
	wg.Add(1)
	go startProxyServerSimple(&wg)

	// 等待所有服务器
	wg.Wait()
}

// 简单版本的目标服务器函数
func startTargetServer1Simple(wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器1的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	log.Println("目标服务器1启动在端口8081")
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func startTargetServer2Simple(wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器2(静态资源)的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	log.Println("目标服务器2启动在端口8082")
	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func startTargetServer3Simple(wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "来自服务器3(默认服务)的响应 - 路径: %s\n时间: %s\n", r.URL.Path, time.Now().Format("2006-01-02 15:04:05"))
	})

	log.Println("目标服务器3启动在端口8083")
	server := &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func startProxyServerSimple(wg *sync.WaitGroup) {
	defer wg.Done()

	// 等待目标服务器启动
	time.Sleep(2 * time.Second)

	// 方式1: 单一目标代理
	fmt.Println("=== 单一目标代理示例 ===")
	singleProxy, err := NewProxyServer("http://localhost:8081")
	if err != nil {
		log.Fatal("创建代理失败:", err)
	}

	// 方式2: 多目标代理
	fmt.Println("=== 多目标代理示例 ===")
	multiProxy := NewMultiProxyServer()
	multiProxy.AddRoute("/api", "http://localhost:8081")
	multiProxy.AddRoute("/static", "http://localhost:8082")
	multiProxy.AddRoute("/", "http://localhost:8083") // 默认路由

	// 启动代理服务器
	mux := http.NewServeMux()

	// 单一代理路由
	mux.Handle("/single/", http.StripPrefix("/single", singleProxy))

	// 多目标代理路由
	mux.Handle("/", multiProxy)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Println("代理服务器启动在端口 8080")
	fmt.Println("单一代理: http://localhost:8080/single/")
	fmt.Println("多目标代理:")
	fmt.Println("  /api -> http://localhost:8081")
	fmt.Println("  /static -> http://localhost:8082")
	fmt.Println("  / -> http://localhost:8083")
	fmt.Println("\n测试URL:")
	fmt.Println("  http://localhost:8080/api/test")
	fmt.Println("  http://localhost:8080/static/css/style.css")
	fmt.Println("  http://localhost:8080/home")
	fmt.Println("  http://localhost:8080/single/target1/test")

	log.Fatal(server.ListenAndServe())
}

// 优雅关闭版本
func mainWithGracefulShutdown() {
	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动所有服务器
	var wg sync.WaitGroup

	// 启动目标服务器
	wg.Add(1)
	go startTargetServer1(ctx, &wg)

	wg.Add(1)
	go startTargetServer2(ctx, &wg)

	wg.Add(1)
	go startTargetServer3(ctx, &wg)

	wg.Add(1)
	go startProxyServer(ctx, &wg)

	// 等待中断信号
	<-sigChan
	log.Println("收到关闭信号，正在优雅关闭...")

	// 取消上下文
	cancel()

	// 等待所有服务器关闭
	wg.Wait()
	log.Println("所有服务器已关闭")
}

// 主函数 - 你可以选择运行简单版本或优雅关闭版本
func main() {
	// 选择运行模式
	if len(os.Args) > 1 && os.Args[1] == "graceful" {
		mainWithGracefulShutdown()
	} else {
		mainSimple()
	}
}
