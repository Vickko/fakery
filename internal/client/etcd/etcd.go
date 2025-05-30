package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"fakery/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient etcd 客户端封装
type EtcdClient struct {
	client *clientv3.Client
	config *conf.Server_Etcd
	logger *log.Helper
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Addr     string            `json:"addr"`
	Metadata map[string]string `json:"metadata"`
}

// NewEtcdClient 创建 etcd 客户端
func NewEtcdClient(serverConf *conf.Server, logger log.Logger) (*EtcdClient, func(), error) {
	helper := log.NewHelper(logger)

	// Get the Etcd specific config from the Server config
	c := serverConf.GetEtcd()

	if c == nil {
		helper.Info("etcd config is nil, skipping etcd client initialization")
		return nil, func() {}, nil
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   c.Endpoints,
		DialTimeout: c.Timeout.AsDuration(),
	})
	if err != nil {
		helper.Errorf("failed to create etcd client: %v", err)
		helper.Warn("etcd client creation failed, continuing without service registration")
		return nil, func() {}, nil
	}

	etcdClient := &EtcdClient{
		client: client,
		config: c,
		logger: helper,
	}

	cleanup := func() {
		helper.Info("closing etcd client")
		if err := client.Close(); err != nil {
			helper.Errorf("failed to close etcd client: %v", err)
		}
	}

	return etcdClient, cleanup, nil
}

// RegisterService 注册服务
func (e *EtcdClient) RegisterService(ctx context.Context, addr string) error {
	if e == nil || e.client == nil {
		return nil
	}

	serviceInfo := ServiceInfo{
		Name:     e.config.Registry.Service.Name,
		Version:  e.config.Registry.Service.Version,
		Addr:     addr,
		Metadata: e.config.Registry.Service.Metadata,
	}

	data, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	key := fmt.Sprintf("%s/%s/%s", e.config.Registry.Namespace, serviceInfo.Name, addr)

	// 为etcd操作创建带超时的context
	operationCtx, cancel := context.WithTimeout(ctx, e.config.Timeout.AsDuration())
	defer cancel()

	// 创建租约
	lease, err := e.client.Grant(operationCtx, int64(e.config.Registry.Ttl.AsDuration().Seconds()))
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}
	// e.logger.Infof("Lease created with ID: %x", lease.ID)

	// 注册服务
	_, err = e.client.Put(operationCtx, key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 续租 - 使用长期context，但添加错误处理
	ch, err := e.client.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		return fmt.Errorf("failed to keep alive lease: %w", err)
	}

	// 启动续租协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				e.logger.Errorf("etcd keep-alive goroutine panic: %v", r)
			}
		}()

		for ka := range ch {
			if ka != nil {
				e.logger.Debugf("lease keep alive response: %v", ka)
			} else {
				e.logger.Warn("received nil keep alive response")
				break
			}
		}
	}()

	e.logger.Infof("service registered: %s", key)
	return nil
}

// UnregisterService 注销服务
func (e *EtcdClient) UnregisterService(ctx context.Context, addr string) error {
	if e == nil || e.client == nil {
		return nil
	}

	key := fmt.Sprintf("%s/%s/%s", e.config.Registry.Namespace, e.config.Registry.Service.Name, addr)

	// 为etcd操作创建带超时的context
	operationCtx, cancel := context.WithTimeout(ctx, e.config.Timeout.AsDuration())
	defer cancel()

	_, err := e.client.Delete(operationCtx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	e.logger.Infof("service unregistered: %s", key)
	return nil
}

// WatchConfig 监听配置变化
func (e *EtcdClient) WatchConfig(ctx context.Context, key string, callback func([]byte)) error {
	if e == nil || e.client == nil {
		return nil
	}

	configKey := fmt.Sprintf("%s/%s", e.config.Config.Namespace, key)

	// 先获取当前配置
	resp, err := e.client.Get(ctx, configKey)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if len(resp.Kvs) > 0 {
		callback(resp.Kvs[0].Value)
	}

	// 监听配置变化
	watchCh := e.client.Watch(ctx, configKey)
	go func() {
		for watchResp := range watchCh {
			for _, event := range watchResp.Events {
				if event.Type == clientv3.EventTypePut {
					callback(event.Kv.Value)
				}
			}
		}
	}()

	e.logger.Infof("watching config: %s", configKey)
	return nil
}

// PutConfig 设置配置
func (e *EtcdClient) PutConfig(ctx context.Context, key string, value []byte) error {
	if e == nil || e.client == nil {
		return nil
	}

	configKey := fmt.Sprintf("%s/%s", e.config.Config.Namespace, key)

	_, err := e.client.Put(ctx, configKey, string(value))
	if err != nil {
		return fmt.Errorf("failed to put config: %w", err)
	}

	e.logger.Infof("config set successfully: %s", configKey)
	return nil
}

// GetConfig 获取配置
func (e *EtcdClient) GetConfig(ctx context.Context, key string) ([]byte, error) {
	if e == nil || e.client == nil {
		return nil, nil
	}

	configKey := fmt.Sprintf("%s/%s", e.config.Config.Namespace, key)

	resp, err := e.client.Get(ctx, configKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}
