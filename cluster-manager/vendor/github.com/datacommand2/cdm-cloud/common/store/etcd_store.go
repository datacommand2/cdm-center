package store

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/micro/go-micro/v2/registry"
	"reflect"
	"sort"
	"strings"
	"time"
)

var (
	// defaultTimeout 옵션에서 Timeout 값이 설정되지 않을 경우 사용하는 디폴트값
	defaultTimeout = 10 * time.Second
)

// etcdStore ETCD 를 사용하기 위한 구조체
type etcdStore struct {
	// options Store 의 옵션
	options Options

	// client ETCD client 구조체
	client *clientv3.Client

	// config ETCD client 설정 값을 위한 config 구조체
	config *clientv3.Config
}

// NewStore etcdStore 생성
func NewStore(name string, opts ...Option) Store {
	c := &clientv3.Config{DialTimeout: defaultTimeout}
	option := Options{
		ServiceName:       name,
		Context:           context.Background(),
		HeartBeatInterval: defaultTimeout,
	}
	for _, o := range opts {
		o(&option)
	}

	e := &etcdStore{
		options: option,
		config:  c,
	}
	e.applyConfig()

	return e
}

// Close etcdStore.client 의 연결 종료
func (e *etcdStore) Close() error {
	return e.client.Close()
}

func (e *etcdStore) getServiceEndpoints() ([]string, error) {
	var r registry.Registry
	if e.options.Registry == nil {
		r = registry.DefaultRegistry
	} else {
		r = e.options.Registry
	}

	s, err := r.GetService(e.options.ServiceName)
	switch {
	case err != nil && err != registry.ErrNotFound:
		logger.Errorf("Could not discover store(%s) service. cause: %v", e.options.ServiceName, err)
		return nil, errors.New("could not discover store service")

	case err == registry.ErrNotFound || s == nil || len(s) == 0:
		logger.Errorf("Not found store(%s) service.", e.options.ServiceName)
		return nil, errors.New("not found store service")
	}

	var endpoints []string
	for _, n := range s[0].Nodes {
		endpoints = append(endpoints, n.Address)
	}

	return endpoints, nil
}

func (e *etcdStore) resetEndpoints() {
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

// Connect etcdStore 에 지정된 옵션으로 etcd client 생성
func (e *etcdStore) Connect() error {
	var err error
	if e.config.Endpoints, err = e.getServiceEndpoints(); err != nil {
		return err
	}

	e.client, err = clientv3.New(*e.config)
	if err != nil {
		logger.Errorf("etcd client creation failed. cause: %v", err)
		return err
	}

	go e.resetEndpoints()

	return nil
}

// Options 현재 설정된 etcdStore 의 옵션을 반환
func (e *etcdStore) Options() Options {
	return e.options
}

// Get key 에 해당하는 value 를 조회
// 옵션으로 timeout 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) Get(key string, opts ...GetOption) (string, error) {
	options := GetOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	resp, err := e.client.KV.Get(ctx, key)
	switch {
	case err == nil && resp.Count == 0:
		err = ErrNotFoundKey
		fallthrough

	case err != nil:
		return "", err

	default:
		return string(resp.Kvs[0].Value), nil
	}
}

// GetAll key 에 해당하는 모든 value 를 조회
// 옵션으로 timeout 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
// json 형태의 string 을 return
func (e *etcdStore) GetAll(key string, opts ...GetOption) (string, error) {
	options := GetOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	resp, err := e.client.KV.Get(ctx, key, clientv3.WithPrefix())
	switch {
	case err == nil && resp.Count == 0:
		err = ErrNotFoundKey
		fallthrough

	case err != nil:
		return "", err

	default:
		data := make(map[string]interface{})
		for _, kv := range resp.Kvs {
			// kv.value 에 있는 모든 \,"{,}" 문자는 삭제
			value := strings.ReplaceAll(string(kv.Value), "\\", "")
			value = strings.ReplaceAll(value, "\"{", "{")
			value = strings.ReplaceAll(value, "}\"", "}")
			data[string(kv.Key)] = value
		}

		rsp, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		return string(rsp), nil
	}
}

// Put key, value 를 저장
// 옵션으로 timeout, TTL 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) Put(key, value string, opts ...PutOption) error {
	options := PutOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	var etcdOpts []clientv3.OpOption
	if options.TTL != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
		defer cancel()

		lr, err := e.client.Lease.Grant(ctx, int64(options.TTL.Seconds()))
		if err != nil {
			return err
		}

		etcdOpts = append(etcdOpts, clientv3.WithLease(lr.ID))
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	_, err := e.client.KV.Put(ctx, key, value, etcdOpts...)

	return err
}

// Delete key 로 해당 key, value 를 삭제
// 옵션으로 timeout, Prefix 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
// prefix 옵션값으로 지정되는 경우 prefix 에 일치하는 모든 key, value 를 삭제함
func (e *etcdStore) Delete(key string, opts ...DeleteOption) error {
	options := DeleteOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	var etcdOpts []clientv3.OpOption
	if options.Prefix {
		etcdOpts = append(etcdOpts, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	_, err := e.client.KV.Delete(ctx, key, etcdOpts...)

	return err
}

// List key 를 prefix 로 하여 일치하는 모든 key 를 조회
// 옵션으로 timeout 이 설정 가능함
// timeout 설정되지 않는경우 디폴트로 설정됨
func (e *etcdStore) List(key string, opts ...ListOption) ([]string, error) {
	options := ListOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	resp, err := e.client.KV.Get(ctx, key, clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	switch {
	case err == nil && resp.Count == 0:
		err = ErrNotFoundKey
		fallthrough

	case err != nil:
		return nil, err

	default:
		var keys []string
		for _, kv := range resp.Kvs {
			keys = append(keys, string(kv.Key))
		}
		return keys, nil
	}
}

// Transaction start a transaction as a block,
// return error will rollback, otherwise to commit.
func (e *etcdStore) Transaction(fc func(Txn) error, opts ...TxnOption) error {
	var options = TxnOptions{
		Timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(&options)
	}

	var txn etcdTxn
	if err := fc(&txn); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	_, err := e.client.KV.Txn(ctx).Then(txn.ops...).Commit()
	cancel()

	return err
}
