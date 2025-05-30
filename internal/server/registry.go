package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"fakery/internal/client/etcd"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Registry 服务注册器
type Registry struct {
	etcdClient *etcd.EtcdClient
	logger     *log.Helper
}

// NewRegistry 创建服务注册器
func NewRegistry(etcdClient *etcd.EtcdClient, logger log.Logger) *Registry {
	return &Registry{
		etcdClient: etcdClient,
		logger:     log.NewHelper(logger),
	}
}

// RegisterServers 注册服务器（并行注册）
func (r *Registry) RegisterServers(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server) error {

	if r.etcdClient == nil {
		r.logger.Info("etcd client is nil, skipping service registration")
		return nil
	}

	var wg sync.WaitGroup

	// 并行注册 HTTP 服务
	if httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			endpoint, err := httpServer.Endpoint()
			if err != nil {
				r.logger.Errorf("failed to get HTTP server endpoint: %v", err)
				return
			}

			httpAddr := r.getServerAddr(endpoint.String())
			r.logger.Infof("Registering HTTP server endpoint to etcd: %s -> %s", endpoint.String(), httpAddr)

			if err := r.etcdClient.RegisterService(ctx, fmt.Sprintf("http://%s", httpAddr)); err != nil {
				r.logger.Errorf("failed to register HTTP service: %v", err)
				r.logger.Warn("HTTP service registration failed, but continuing with startup")
			}
		}()
	}

	// 并行注册 gRPC 服务
	if grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			endpoint, err := grpcServer.Endpoint()
			if err != nil {
				r.logger.Errorf("failed to get gRPC server endpoint: %v", err)
				return
			}

			grpcAddr := r.getServerAddr(endpoint.String())
			r.logger.Infof("Registering gRPC server endpoint to etcd: %s -> %s", endpoint.String(), grpcAddr)

			if err := r.etcdClient.RegisterService(ctx, fmt.Sprintf("grpc://%s", grpcAddr)); err != nil {
				r.logger.Errorf("failed to register gRPC service: %v", err)
				r.logger.Warn("gRPC service registration failed, but continuing with startup")
			}
		}()
	}

	// 等待所有注册完成
	wg.Wait()
	r.logger.Info("Service registration completed")
	return nil
}

// UnregisterServers 注销服务器（并行注销）
func (r *Registry) UnregisterServers(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server) error {
	if r.etcdClient == nil {
		return nil
	}

	var wg sync.WaitGroup

	// 并行注销 HTTP 服务
	if httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.logger.Info("Unregistering HTTP service on etcd...")

			endpoint, err := httpServer.Endpoint()
			if err != nil {
				r.logger.Errorf("failed to get HTTP server endpoint: %v", err)
				return
			}

			httpAddr := r.getServerAddr(endpoint.String())
			if err := r.etcdClient.UnregisterService(ctx, fmt.Sprintf("http://%s", httpAddr)); err != nil {
				r.logger.Errorf("failed to unregister HTTP service: %v", err)
			}
		}()
	}

	// 并行注销 gRPC 服务
	if grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.logger.Info("Unregistering gRPC service...")

			endpoint, err := grpcServer.Endpoint()
			if err != nil {
				r.logger.Errorf("failed to get gRPC server endpoint: %v", err)
				return
			}

			grpcAddr := r.getServerAddr(endpoint.String())
			if err := r.etcdClient.UnregisterService(ctx, fmt.Sprintf("grpc://%s", grpcAddr)); err != nil {
				r.logger.Errorf("failed to unregister gRPC service: %v", err)
			}
		}()
	}

	// 等待所有注销完成
	wg.Wait()
	r.logger.Info("Service unregistration completed")

	return nil
}

// getServerAddr 获取服务器地址
func (r *Registry) getServerAddr(endpoint string) string {
	if endpoint == "" {
		return ""
	}

	// 解析地址
	if host, port, err := net.SplitHostPort(endpoint); err == nil {
		// 如果 host 是 0.0.0.0 或空，使用本机 IP
		if host == "0.0.0.0" || host == "" {
			if localIP := r.getLocalIP(); localIP != "" {
				return net.JoinHostPort(localIP, port)
			}
		}
		return endpoint
	}

	return endpoint
}

// getLocalIP 获取本机 IP
func (r *Registry) getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
