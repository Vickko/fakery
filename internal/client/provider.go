package client

import (
	"github.com/google/wire"

	"fakery/internal/client/etcd"
)

// ProviderSet is etcd client providers.
var ProviderSet = wire.NewSet(etcd.NewEtcdClient)
