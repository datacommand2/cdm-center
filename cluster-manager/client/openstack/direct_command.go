package openstack

import (
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
)

type directCEPH struct {
	clusterID   uint64
	storageID   uint64
	metadata    map[string]interface{}
	backendName string
}

func newDirectCEPH(tgtStorage model.ClusterStorage, metadata map[string]string) (*directCEPH, error) {
	backendName := metadata["backend"]

	md := make(map[string]interface{})
	for k, v := range metadata {
		md[k] = v
	}

	return &directCEPH{
		clusterID:   tgtStorage.ClusterID,
		storageID:   tgtStorage.ID,
		metadata:    md,
		backendName: backendName,
	}, nil
}

func cephVolumeName(volUUID string) string {
	return fmt.Sprintf("volume-%s", volUUID)
}

func cephSnapshotName(snpUUID string) string {
	return fmt.Sprintf("snapshot-%s", snpUUID)
}

func (c *directCEPH) getBackendName() string {
	return c.backendName
}
