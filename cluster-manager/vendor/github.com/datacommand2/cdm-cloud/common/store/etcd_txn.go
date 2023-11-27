package store

import (
	"github.com/coreos/etcd/clientv3"
)

type etcdTxn struct {
	ops []clientv3.Op
}

// Put key, value 를 저장
func (e *etcdTxn) Put(key, value string) {
	e.ops = append(e.ops, clientv3.OpPut(key, value))
}

// Delete key 로 해당 key, value 를 삭제
func (e *etcdTxn) Delete(key string, opts ...DeleteOption) {
	options := DeleteOptions{}
	for _, o := range opts {
		o(&options)
	}

	var etcdOpts []clientv3.OpOption
	if options.Prefix {
		etcdOpts = append(etcdOpts, clientv3.WithPrefix())
	}

	e.ops = append(e.ops, clientv3.OpDelete(key, etcdOpts...))
}
