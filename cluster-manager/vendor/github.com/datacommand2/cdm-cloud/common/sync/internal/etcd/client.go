package etcd

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/micro/go-micro/v2/registry"
	"reflect"
	"sort"
	"time"
)

var (
	// defaultTimeout 옵션 에서 Timeout 값이 설정 되지 않을 경우 사용 하는 디폴트 값
	defaultTimeout = 10 * time.Second
)

// etcdClient ETCD 를 사용 하기 위한 etcd client 구조체
type etcdClient struct {
	// options 옵션
	options options

	// client ETCD client 구조체
	client *clientv3.Client

	// config ETCD client 설정 값을 위한 config 구조체
	config *clientv3.Config
}

// NewClient etcd client 생성
func NewClient(name string, opts ...Option) (*clientv3.Client, error) {
	c := &clientv3.Config{DialTimeout: defaultTimeout}
	option := options{
		ServiceName: name,
		Context:     context.Background(),
	}
	for _, o := range opts {
		o(&option)
	}

	cli, err := (&etcdClient{options: option, config: c}).init()
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (e *etcdClient) getServiceEndpoints() ([]string, error) {
	var r registry.Registry
	if e.options.Registry == nil {
		r = registry.DefaultRegistry
	} else {
		r = e.options.Registry
	}

	s, err := r.GetService(e.options.ServiceName)
	switch {
	case err != nil && err != registry.ErrNotFound:
		logger.Errorf("Could not discover (%s)service. cause: %v", e.options.ServiceName, err)
		return nil, errors.New("could not discover service")

	case err == registry.ErrNotFound || s == nil || len(s) == 0:
		logger.Errorf("Not found (%s) service.", e.options.ServiceName)
		return nil, errors.New("not found service")
	}

	var endpoints []string
	for _, n := range s[0].Nodes {
		endpoints = append(endpoints, n.Address)
	}

	return endpoints, nil
}

func (e *etcdClient) resetEndpoints() {
	for {
		time.Sleep(e.options.HeartBeatInterval)

		ep1, _ := e.getServiceEndpoints()
		ep2 := e.client.Endpoints()
		sort.Strings(ep1)
		sort.Strings(ep2)

		if reflect.DeepEqual(ep1, ep2) {
			continue
		}

		if len(ep1) != 0 {
			e.client.SetEndpoints(ep1...)
		}
	}
}

func (e *etcdClient) init() (*clientv3.Client, error) {
	e.applyConfig()

	var err error
	if e.config.Endpoints, err = e.getServiceEndpoints(); err != nil {
		return nil, err
	}

	e.client, err = clientv3.New(*e.config)
	if err != nil {
		logger.Errorf("etcdClient client creation failed. cause: %v", err)
		return nil, err
	}

	go e.resetEndpoints()

	return e.client, err
}
