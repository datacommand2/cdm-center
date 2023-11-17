package sync

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"time"
)

func publishMessage(topic string, obj interface{}) error {
	if topic == "" {
		return nil
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return broker.Publish(topic, &broker.Message{Body: b})
}

func waitVolumeStatus(cli client.Client, uuid string) error {
	timeout := time.After(120 * time.Second)
	for {
		rsp, err := cli.GetVolume(client.GetVolumeRequest{Volume: model.ClusterVolume{UUID: uuid}})
		if errors.Equal(err, client.ErrNotFound) {
			return nil
		} else if errors.Equal(err, openstack.ErrNotFoundVolumeType) {
			return nil
		} else if err != nil {
			return err
		}

		switch rsp.Result.Volume.Status {
		case "attaching", "deleting", "extending", "unmanaging", "detaching", "retyping":
			break
		default:
			return nil
		}

		select {
		case <-timeout:
			return VolumeStatusTimeout(rsp.Result.Volume.Status, uuid)

		case <-time.After(3 * time.Second):
			continue
		}
	}
}

func waitInstanceStatus(cli client.Client, uuid string) error {
	timeout := time.After(60 * time.Second)
	for {
		rsp, err := cli.GetInstance(client.GetInstanceRequest{Instance: model.ClusterInstance{UUID: uuid}})
		if errors.Equal(err, client.ErrNotFound) {
			return nil
		} else if errors.Equal(err, openstack.ErrNotFoundInstanceHypervisor) {
			return nil
		} else if err != nil {
			return err
		}

		if rsp.Result.Instance.Status != "MIGRATING" {
			return nil
		}

		select {
		case <-timeout:
			return InstanceStatusTimeout(rsp.Result.Instance.Status, uuid)

		case <-time.After(3 * time.Second):
			continue
		}
	}
}
