package leader

import "github.com/datacommand2/cdm-center/cluster-manager/queue"

// Options sync 시 사용하는 옵션 구조체
type Options struct {
	OpenStack    bool
	Excluded     *queue.ExcludedCluster
	CheckCluster *queue.ClusterServiceStatus
}

// Option 옵션 함수
type Option func(*Options)

// LoadDataFromOpenStack CheckCluster 에서 최신 데이터를 불러오는 OpenStack API 사용여부를 결정하는 옵션 함수
func LoadDataFromOpenStack(l bool) Option {
	return func(o *Options) {
		o.OpenStack = l
	}
}

// PutCheckCluster 함수는 있는 CheckCluster 데이터를 넣어 사용한다.
func PutCheckCluster(c queue.ClusterServiceStatus) Option {
	return func(o *Options) {
		o.CheckCluster = &queue.ClusterServiceStatus{
			ClusterID:    c.ClusterID,
			Storage:      c.Storage,
			Compute:      c.Compute,
			Network:      c.Network,
			Status:       c.Status,
			StorageError: c.StorageError,
			ComputeError: c.ComputeError,
			NetworkError: c.NetworkError,
		}
	}
}

// PutExcluded 함수는 가지고 있는 Excluded 데이터를 넣어 사용한다.
func PutExcluded(c *queue.ExcludedCluster) Option {
	return func(o *Options) {
		o.Excluded = &queue.ExcludedCluster{
			ClusterID: c.ClusterID,
			Storage:   c.Storage,
			Compute:   c.Compute,
			Network:   c.Network,
		}
	}
}

func newOptions(opts ...Option) *Options {
	opt := &Options{
		OpenStack:    false,
		CheckCluster: nil,
	}

	for _, o := range opts {
		o(opt)
	}

	return opt
}
