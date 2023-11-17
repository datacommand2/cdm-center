package notification

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
)

func getInstancesFromVolume(volumeUUID string) ([]*model.ClusterInstance, error) {
	var instances []*model.ClusterInstance

	if err := database.Transaction(func(db *gorm.DB) error {
		return db.Table(model.ClusterInstance{}.TableName()).
			Joins("join cdm_cluster_instance_volume on cdm_cluster_instance_volume.cluster_instance_id = cdm_cluster_instance.id").
			Joins("join cdm_cluster_volume on cdm_cluster_volume.id = cdm_cluster_instance_volume.cluster_volume_id").
			Where(&model.ClusterVolume{UUID: volumeUUID}).
			Find(&instances).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return instances, nil
}

func getInstanceFromFloatingIP(clusterID uint64, payload map[string]interface{}) (*model.ClusterInstance, error) {
	var instance model.ClusterInstance
	var err error

	if payload["port_id"] != nil {
		var c model.Cluster

		if err := database.Transaction(func(db *gorm.DB) error {
			return db.First(&c, model.Cluster{ID: clusterID}).Error
		}); err != nil {
			return nil, errors.UnusableDatabase(err)
		}

		c.Credential, err = client.DecryptCredentialPassword(c.Credential)
		if err != nil {
			logger.Errorf("Could not update state of cluster(%d:%s). Cause: %+v", c.ID, c.Name, err)
		}

		cli, err := client.New(c.TypeCode, c.APIServerURL, c.Credential, "")
		if err != nil {
			return nil, err
		}

		if err := cli.Connect(); err != nil {
			logger.Errorf("[getInstanceFromFloatingIP] Could not connect. Cause: %v", err)
			return nil, err
		}

		defer func() {
			_ = cli.Close()
		}()

		rsp, err := train.ShowPortDetails(cli, train.ShowPortDetailsRequest{PortID: payload["port_id"].(string)})
		if err != nil {
			return nil, err
		}

		if rsp.Port.DeviceOwner != "compute:nova" {
			return nil, nil
		}

		instance.UUID = rsp.Port.DeviceID

	} else {

		err := database.Transaction(func(db *gorm.DB) error {
			return db.Table(model.ClusterInstance{}.TableName()).
				Joins("join cdm_cluster_instance_network on cdm_cluster_instance_network.cluster_instance_id = cdm_cluster_instance.id").
				Joins("join cdm_cluster_floating_ip on cdm_cluster_floating_ip.id = cdm_cluster_instance_network.cluster_floating_ip_id").
				Where(&model.ClusterFloatingIP{UUID: payload["id"].(string)}).
				First(&instance).Error
		})
		switch {
		case err == gorm.ErrRecordNotFound:
			return nil, nil

		case err != nil:
			return nil, errors.UnusableDatabase(err)
		}
	}

	return &instance, nil
}
