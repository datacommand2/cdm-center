package handler

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	clusterStorage "github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-center/cluster-manager/volume"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/datacommand2/cdm-cloud/common/util"
)

func createError(ctx context.Context, eventCode string, err error) error {
	if err == nil {
		return nil
	}

	var errorCode string
	switch {
	// openstack
	case errors.Equal(err, openstack.ErrNotFoundVolumeType):
		errorCode = "not_found_volume_type"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundVolumeBackendName):
		errorCode = "not_found_volume_backend_name"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundAvailableServiceHostName):
		errorCode = "not_found_available_service_host_name"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundVolumeSourceName):
		errorCode = "not_found_volume_source_name"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundSnapshotSourceName):
		errorCode = "not_found_snapshot_source_name"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundInstanceHypervisor):
		errorCode = "not_found_instance_hypervisor"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundInstanceSpec):
		errorCode = "not_found_instance_spec"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrUnknownInstanceState):
		errorCode = "unknown_instance_state"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrImportVolumeRollbackFailed):
		errorCode = "import_volume_rollback_failed"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundNFSExportPath):
		errorCode = "not_found_NFS_export_path"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrUnusableMirrorVolume):
		errorCode = "unusable_mirror_volume"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, openstack.ErrNotFoundManagedVolume):
		errorCode = "not_found_managed_volume"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	// client
	case errors.Equal(err, client.ErrNotFound):
		errorCode = "not_found"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrUnAuthenticated):
		errorCode = "unauthenticated"
		return errors.StatusUnauthenticated(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrBadRequest):
		errorCode = "bad_request"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrUnAuthorized):
		errorCode = "unauthorized"
		return errors.StatusUnauthorized(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrUnAuthorizedUser):
		errorCode = "unauthorized_user"
		return errors.StatusUnauthorized(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrNotConnected):
		errorCode = "not_connected"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrConflict):
		errorCode = "conflict"
		return errors.StatusConflict(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrNotFoundEndpoint):
		errorCode = "not_found_endpoint"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrUnknown):
		errorCode = "unknown"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, client.ErrRemoteServerError):
		errorCode = "remote_server_error"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	// internal
	case errors.Equal(err, internal.ErrNotFoundGroup):
		errorCode = "not_found_group"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundCluster):
		errorCode = "not_found_cluster"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterHypervisor):
		errorCode = "not_found_cluster_hypervisor"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterNetwork):
		errorCode = "not_found_cluster_network"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterNetworkByUUID):
		errorCode = "not_found_cluster_network_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterSubnet):
		errorCode = "not_found_cluster_subnet"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterFloatingIP):
		errorCode = "not_found_cluster_floating_ip"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterTenant):
		errorCode = "not_found_cluster_tenant"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterTenantByUUID):
		errorCode = "not_found_cluster_tenant_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterInstance):
		errorCode = "not_found_cluster_instance"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterInstanceByUUID):
		errorCode = "not_found_cluster_instance_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterInstanceSpec):
		errorCode = "not_found_cluster_instance_spec"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterInstanceSpecByUUID):
		errorCode = "not_found_cluster_instance_spec_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterKeyPair):
		errorCode = "not_found_cluster_keypair"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrUnsupportedClusterType):
		errorCode = "unsupported_cluster_type"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterAvailabilityZone):
		errorCode = "not_found_cluster_availability_zone"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterVolume):
		errorCode = "not_found_cluster_volume"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterVolumeByUUID):
		errorCode = "not_found_cluster_volume_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterStorage):
		errorCode = "not_found_cluster_storage"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundRouterExternalNetwork):
		errorCode = "not_found_router_external_network"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterRouter):
		errorCode = "not_found_cluster_router"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterRouterByUUID):
		errorCode = "not_found_cluster_router_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterSecurityGroup):
		errorCode = "not_found_cluster_security_group"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterSecurityGroupByUUID):
		errorCode = "not_found_cluster_security_group_by_uuid"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterSyncStatus):
		errorCode = "not_found_cluster_sync_status"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrCurrentPasswordMismatch):
		errorCode = "current_password_mismatch"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrUnableSynchronizeStatus):
		errorCode = "unable_to_synchronize_because_of_status"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, internal.ErrNotFoundClusterException):
		errorCode = "not_found_cluster_sync_exception"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	// cluster storage
	case errors.Equal(err, clusterStorage.ErrNotFoundStorage):
		errorCode = "not_found_storage"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, clusterStorage.ErrUnsupportedStorageType):
		errorCode = "unsupported_storage_type"
		return errors.StatusBadRequest(ctx, eventCode, errorCode, err)

	case errors.Equal(err, clusterStorage.ErrNotFoundStorageMetadata):
		errorCode = "not_found_storage_metadata"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	// sync
	case errors.Equal(err, sync.ErrVolumeStatusTimeout):
		errorCode = "volume_status_timeout"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	case errors.Equal(err, sync.ErrInstanceStatusTimeout):
		errorCode = "instance_status_timeout"
		return errors.StatusInternalServerError(ctx, eventCode, errorCode, err)

	// volume
	case errors.Equal(err, volume.ErrNotFoundVolume):
		errorCode = "not_found_volume"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	case errors.Equal(err, volume.ErrNotFoundVolumeMetadata):
		errorCode = "not_found_volume_metadata"
		return errors.StatusNotFound(ctx, eventCode, errorCode, err)

	default:
		if err := util.CreateError(ctx, eventCode, err); err != nil {
			return err
		}
	}

	return nil
}

func validateRequest(ctx context.Context) error {
	if _, err := metadata.GetTenantID(ctx); err != nil {
		return errors.InvalidRequest(ctx)
	}

	if _, err := metadata.GetAuthenticatedUser(ctx); err != nil {
		return errors.InvalidRequest(ctx)
	}

	return nil
}
