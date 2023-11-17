package notification

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net"
	"net/url"
	"strings"
)

func init() {
	client.RegisterClusterMonitorCreationFunc(constant.ClusterTypeOpenstack, New)
}

type notificationHandler struct {
	cluster *model.Cluster
}

// Start openstack notification Start
func (n *notification) Start(cluster *model.Cluster) error {
	if err := n.Connect(); err != nil {
		return errors.UnusableBroker(err)
	}

	if err := n.DeleteQueue(); err != nil {
		logger.Warnf("Could not delete openstack notification queue. Cause: %+v", err)
	}

	handler := notificationHandler{
		cluster: cluster,
	}

	if _, err := n.Subscribe(handler.HandleEvent); err != nil {
		_ = n.Disconnect()
		return errors.UnusableBroker(err)
	}

	return nil
}

// Stop openstack notification stop
func (n *notification) Stop() {
	_ = n.Disconnect()
}

func (ns *notificationHandler) HandleEvent(p Event) error {
	var e map[string]interface{}
	err := json.Unmarshal(p.Message().Body, &e)
	if err != nil {
		logger.Errorf("[Notification] Could not handle event. Cause: %+v", err)
		return err
	}

	var m OsloMessage
	if err := json.Unmarshal([]byte(e["oslo.message"].(string)), &m); err != nil {
		logger.Errorf("[Notification] Could not handle event. Cause: %+v", err)
		return err
	}

	var uuid string
	var tenantUUID string
	switch m.EventType {
	case projectCreated, projectUpdated, projectDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["resource_info"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceTenant, uuid)

	case projectRoleAssignmentCreated, projectRoleAssignmentDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["project"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceTenant, uuid)

	case instanceCreated, instanceUpdated, instanceDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["instance_id"].(string)

		// instanceCreated, instanceUpdated
		if m.EventType != instanceDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceInstance, uuid)
			return nil
		}

		// instanceDeleted
		tenantUUID = m.Payload["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceInstance, uuid, sync.TenantUUID(tenantUUID))

	case volumeCreated, volumeUpdated, volumeResetStatus, volumeManaged, volumeUnmanaged, volumeResized, volumeRetyped, volumeDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["volume_id"].(string)

		if m.EventType != volumeDeleted {
			// volume retype 시 openstack 에서 만들어진 임시 볼륨의 삭제에 대한 이벤트가 오지 않기 때문에 cluster-manager 자체적으로
			// 주기적 동기화를 통해 DB에 만들어진 임시 볼륨 정보를 삭제한다.
			sync.Reserve(ns.cluster.ID, sync.ResourceVolume, uuid)
			return nil
		}

		// volumeDeleted
		tenantUUID = m.Payload["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceVolume, uuid, sync.TenantUUID(tenantUUID))

	case volumeAttached, volumeDetached:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["volume_id"].(string)

		if m.EventType != volumeDetached {
			sync.Reserve(ns.cluster.ID, sync.ResourceVolume, uuid)
			for _, attachment := range m.Payload["volume_attachment"].([]interface{}) {
				instance := attachment.(map[string]interface{})["instance_uuid"].(string)
				sync.Reserve(ns.cluster.ID, sync.ResourceInstance, instance)
			}
			return nil
		}

		// volumeDetached
		tenantUUID = m.Payload["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceVolume, uuid, sync.TenantUUID(tenantUUID))
		instances, err := getInstancesFromVolume(uuid)
		if err != nil {
			logger.Errorf("[Notification] Failed to get instance from volume(%s). cause : %+v", uuid, err)
			return err
		}

		for _, instance := range instances {
			sync.Reserve(ns.cluster.ID, sync.ResourceInstance, instance.UUID)
		}

	case volumeSnapshotCreated, volumeSnapshotUpdated, volumeSnapshotResetStatus, volumeSnapshotDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["snapshot_id"].(string)

		if m.EventType != volumeSnapshotDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceVolumeSnapshot, uuid)
			return nil
		}

		// volumeSnapshotDeleted
		tenantUUID = m.Payload["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceVolumeSnapshot, uuid, sync.TenantUUID(tenantUUID))

	case volumeTypeCreated, volumeTypeUpdated, volumeTypeDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["volume_types"].(map[string]interface{})["id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceStorage, uuid)

	case volumeTypeProjectAccessAdd:

	case volumeTypeExtraSpecsCreated, volumeTypeExtraSpecsDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["type_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceStorage, uuid)

	case networkCreated, networkUpdated, networkDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["network"].(map[string]interface{})["id"].(string)

		if m.EventType != networkDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceNetwork, uuid)
			return nil
		}

		// networkDeleted
		tenantUUID = m.Payload["network"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceNetwork, uuid, sync.TenantUUID(tenantUUID))

	case subnetCreated, subnetUpdated, subnetDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["subnet"].(map[string]interface{})["id"].(string)

		if m.EventType != subnetDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceSubnet, uuid)
			return nil
		}

		// subnetDeleted
		tenantUUID = m.Payload["subnet"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceSubnet, uuid, sync.TenantUUID(tenantUUID))

	case securityGroupCreated, securityGroupUpdated, securityGroupDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["security_group"].(map[string]interface{})["id"].(string)

		if m.EventType != securityGroupDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceSecurityGroup, uuid)
			return nil
		}

		// securityGroupDeleted
		tenantUUID = m.Payload["security_group"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceSecurityGroup, uuid, sync.TenantUUID(tenantUUID))

	case securityGroupRuleCreated, securityGroupRuleUpdated, securityGroupRuleDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["security_group_rule"].(map[string]interface{})["id"].(string)

		if m.EventType != securityGroupRuleDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceSecurityGroupRule, uuid)
			return nil
		}

		// securityGroupRuleDeleted
		tenantUUID = m.Payload["security_group_rule"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceSecurityGroupRule, uuid, sync.TenantUUID(tenantUUID))

	case routerCreated, routerUpdated, routerDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["router"].(map[string]interface{})["id"].(string)

		if m.EventType != routerDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceRouter, uuid)
			return nil
		}

		// routerDeleted
		tenantUUID = m.Payload["router"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceRouter, uuid, sync.TenantUUID(tenantUUID))

	case routerInterfaceCreated, routerInterfaceDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["router_interface"].(map[string]interface{})["id"].(string)

		if m.EventType != routerInterfaceDeleted {
			sync.Reserve(ns.cluster.ID, sync.ResourceRouter, uuid)
			return nil
		}

		// routerInterfaceDeleted
		tenantUUID = m.Payload["router_interface"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceRouter, uuid, sync.TenantUUID(tenantUUID))

	case floatingIPCreated, floatingIPUpdated, floatingIPDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["floatingip"].(map[string]interface{})["id"].(string)

		if m.EventType == floatingIPDeleted {
			tenantUUID = m.Payload["floatingip"].(map[string]interface{})["tenant_id"].(string)
			sync.Reserve(ns.cluster.ID, sync.ResourceFloatingIP, uuid, sync.TenantUUID(tenantUUID))
			return nil
		}

		// floatingIPCreated, floatingIPUpdated
		sync.Reserve(ns.cluster.ID, sync.ResourceFloatingIP, uuid)

		if m.EventType == floatingIPUpdated {
			instance, err := getInstanceFromFloatingIP(ns.cluster.ID, m.Payload["floatingip"].(map[string]interface{}))
			if err != nil {
				logger.Errorf("[Notification] Failed get instance from floating ip(%s). cause %+v", uuid, err)
				return err
			}

			if instance != nil {
				sync.Reserve(ns.cluster.ID, sync.ResourceInstance, instance.UUID)
			}
		}

	case portUpdated, portDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["port"].(map[string]interface{})["device_id"].(string)

		// instance attach, detach interface 시 payload 의 정보중 device_owner 가 compute:nova 인 경우에는 device_id로 instance 동기화를 한다.
		// router interface 의 port 수정 시 payload 의 정보중 device_owner 가 network:router_interface 인 경우에는 device_id로 router 동기화를 한다.
		// 나머지 포트 이벤트 알림에 대해선 처리하지 않는다.
		var resource string
		switch m.Payload["port"].(map[string]interface{})["device_owner"].(string) {
		case "compute:nova":
			resource = sync.ResourceInstance
		case "network:router_interface":
			resource = sync.ResourceRouter
		default:
			return nil
		}

		if m.EventType != portDeleted {
			sync.Reserve(ns.cluster.ID, resource, uuid)
			return nil
		}

		// portDeleted
		tenantUUID = m.Payload["port"].(map[string]interface{})["tenant_id"].(string)
		sync.Reserve(ns.cluster.ID, resource, uuid, sync.TenantUUID(tenantUUID))

	case keypairCreated, keypairImported, keypairDeleted:
		if !isExistedKey(m) {
			return nil
		}
		uuid = m.Payload["key_name"].(string)
		sync.Reserve(ns.cluster.ID, sync.ResourceKeypair, uuid)

	case aggregateCreated, aggregatePropUpdated, aggregateMetadataUpdated:
		sync.Reserve(ns.cluster.ID, sync.ResourceAvailabilityZoneList, "")

	case aggregateHostAdded, aggregateHostRemoved:
		sync.Reserve(ns.cluster.ID, sync.ResourceHypervisorList, "")

	case aggregateDeleted:
		sync.Reserve(ns.cluster.ID, sync.ResourceHypervisorList, "")
		sync.Reserve(ns.cluster.ID, sync.ResourceAvailabilityZoneList, "")

	case "telemetry.polling", "",
		"compute.instance.exists", "compute.instance.shutdown.end", "compute_task.build_instances",
		"group_type.create", "group_type.delete",
		"group.create.end", "group.update.end", "group.delete.end",
		"group_snapshot.create.end", "group_snapshot.delete.end",
		"port.create.end", "volume.manage_existing.end", "snapshot.manage_existing_snapshot.end",
		"scheduler.select_destinations.end":
		return nil

	case "compute.instance.create.error":
		logger.Infof("[Notification] Error occurred during creating instance. Payload >> %+v", m)

	default:
		if !strings.Contains(m.EventType, ".start") {
			logger.Warnf("[Notification] Payload is not supported >> %+v", m)
		}
	}

	return nil
}

// New 함수는 새로운 monitor interface 를 생성한다.
func New(serverURL string) (client.Monitor, error) {
	//TODO auth 의 경우 임시로 ID:PASSWORD(ex.guest:guest)를 쓰지만
	//	   사용자 입력에 의한 Cluster 의 MetaData 로 저장될 필요가 있음.
	//     마찬가지로 임시로 client 의 api server url 과 고정된 port(ex.192.168.1.1:5672) 를 쓰지만
	//     사용자 입력에 의한 Cluster 의 MetaData 로 저장될 필요가 있음.
	auth := "guest:guest"
	defaultPort := "5672"

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	ip, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	return &notification{
		auth:    auth,
		address: fmt.Sprintf("%s:%s", ip, defaultPort),
	}, nil
}

func isExistedKey(m OsloMessage) bool {
	switch m.EventType {
	case projectCreated, projectUpdated, projectDeleted:
		value, ok := m.Payload["resource_info"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceTenant(resource_info) event >> %+v", m)
			return false
		}

	case projectRoleAssignmentCreated, projectRoleAssignmentDeleted:
		value, ok := m.Payload["project"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceTenant(project) event >> %+v", m)
			return false
		}

	case instanceCreated, instanceUpdated, instanceDeleted:
		value, ok := m.Payload["instance_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance(instance_id) event >> %+v", m)
			return false
		}

		if m.EventType == instanceDeleted {
			value, ok = m.Payload["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance(tenant_id) event >> %+v", m)
				return false
			}
		}

	case volumeCreated, volumeUpdated, volumeResetStatus, volumeManaged, volumeUnmanaged, volumeResized, volumeRetyped, volumeDeleted:
		value, ok := m.Payload["volume_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolume(volume_id) event >> %+v", m)
			return false
		}

		if m.EventType == volumeDeleted {
			value, ok = m.Payload["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolume(tenant_id) event >> %+v", m)
				return false
			}
		}

	case volumeAttached, volumeDetached:
		value, ok := m.Payload["volume_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolume(volume_id) event >> %+v", m)
			return false
		}

		values, ok := m.Payload["volume_attachment"].([]interface{})
		if !ok {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolume(volume_attachment) event >> %+v", m)
			return false
		}

		for _, v := range values {
			i, ok := v.(map[string]interface{})["instance_uuid"].(string)
			if !ok || i == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance(instance_uuid) event >> %+v", v)
				return false
			}
		}

		if m.EventType == volumeDetached {
			value, ok = m.Payload["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolume(tenant_id) event >> %+v", m)
				return false
			}
		}

	case volumeSnapshotCreated, volumeSnapshotUpdated, volumeSnapshotResetStatus, volumeSnapshotDeleted:
		value, ok := m.Payload["snapshot_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolumeSnapshot(snapshot_id) event >> %+v", m)
			return false
		}

		if m.EventType == volumeSnapshotDeleted {
			value, ok = m.Payload["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceVolumeSnapshot(tenant_id) event >> %+v", m)
				return false
			}
		}

	case volumeTypeCreated, volumeTypeUpdated, volumeTypeDeleted:
		value, ok := m.Payload["volume_types"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceStorage(volume_types.id) event >> %+v", m)
			return false
		}

	case volumeTypeExtraSpecsCreated, volumeTypeExtraSpecsDeleted:
		value, ok := m.Payload["type_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceStorage(type_id) event >> %+v", m)
			return false
		}

	case networkCreated, networkUpdated, networkDeleted:
		value, ok := m.Payload["network"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceNetwork(network) event >> %+v", m)
			return false
		}

		if m.EventType == networkDeleted {
			value, ok = m.Payload["network"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceNetwork(network.tenant_id) event >> %+v", m)
				return false
			}
		}

	case subnetCreated, subnetUpdated, subnetDeleted:
		value, ok := m.Payload["subnet"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSubnet(subnet.id) event >> %+v", m)
			return false
		}

		if m.EventType == subnetDeleted {
			value, ok = m.Payload["subnet"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSubnet(subnet.tenant_id) event >> %+v", m)
				return false
			}
		}

	case securityGroupCreated, securityGroupUpdated, securityGroupDeleted:
		value, ok := m.Payload["security_group"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSecurityGroup(security_group.id) event >> %+v", m)
			return false
		}

		if m.EventType == securityGroupDeleted {
			value, ok = m.Payload["security_group"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSecurityGroup(security_group.tenant_id) event >> %+v", m)
				return false
			}
		}

	case securityGroupRuleCreated, securityGroupRuleUpdated, securityGroupRuleDeleted:
		value, ok := m.Payload["security_group_rule"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSecurityGroupRule(security_group_rule.id) event >> %+v", m)
			return false
		}

		if m.EventType == securityGroupRuleDeleted {
			value, ok = m.Payload["security_group_rule"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceSecurityGroupRule(security_group_rule.tenant_id) event >> %+v", m)
				return false
			}
		}

	case routerCreated, routerUpdated, routerDeleted:
		value, ok := m.Payload["router"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceRouter(router.id) event >> %+v", m)
			return false
		}

		if m.EventType == routerDeleted {
			value, ok = m.Payload["router"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceRouter(router.tenant_id) event >> %+v", m)
				return false
			}
		}

	case routerInterfaceCreated, routerInterfaceDeleted:
		value, ok := m.Payload["router_interface"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceRouter(router_interface.id) event >> %+v", m)
			return false
		}

		if m.EventType == routerInterfaceDeleted {
			value, ok = m.Payload["router_interface"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceRouter(router_interface.tenant_id) event >> %+v", m)
				return false
			}
		}

	case floatingIPCreated, floatingIPUpdated, floatingIPDeleted:
		value, ok := m.Payload["floatingip"].(map[string]interface{})["id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceFloatingIP(floatingip.id) event >> %+v", m)
			return false
		}

		if m.EventType == floatingIPDeleted {
			value, ok = m.Payload["floatingip"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceFloatingIP(floatingip.tenant_id) event >> %+v", m)
				return false
			}
		}

	case portUpdated, portDeleted:
		value, ok := m.Payload["port"].(map[string]interface{})["device_owner"].(string)
		if !ok {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance/ResourceRouter(port.device_owner) event >> %+v", m)
			return false
		} else if value == "" {
			return false
		}

		value, ok = m.Payload["port"].(map[string]interface{})["device_id"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance/ResourceRouter(port.device_id) event >> %+v", m)
			return false
		}

		if m.EventType == portDeleted {
			value, ok = m.Payload["port"].(map[string]interface{})["tenant_id"].(string)
			if !ok || value == "" {
				logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceInstance/ResourceRouter(port.tenant_id) event >> %+v", m)
				return false
			}
		}

	case keypairCreated, keypairImported, keypairDeleted:
		value, ok := m.Payload["key_name"].(string)
		if !ok || value == "" {
			logger.Warnf("[Notification] Payload ERROR(could not sync) - ResourceKeypair(key_name) event >> %+v", m)
			return false
		}
	}

	return true
}
