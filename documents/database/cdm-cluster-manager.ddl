use cdm;

-- table
DROP TABLE IF EXISTS cdm_cluster_routing_interface;
DROP TABLE IF EXISTS cdm_cluster_router_extra_route;
DROP TABLE IF EXISTS cdm_cluster_router;
DROP TABLE IF EXISTS cdm_cluster_subnet_dhcp_pool;
DROP TABLE IF EXISTS cdm_cluster_instance_network;
DROP TABLE IF EXISTS cdm_cluster_subnet_nameserver;
DROP TABLE IF EXISTS cdm_cluster_subnet;
DROP TABLE IF EXISTS cdm_cluster_instance_volume;
DROP TABLE IF EXISTS cdm_cluster_floating_ip;
DROP TABLE IF EXISTS cdm_cluster_network;
DROP TABLE IF EXISTS cdm_cluster_instance_security_group;
DROP TABLE IF EXISTS cdm_cluster_security_group_rule;
DROP TABLE IF EXISTS cdm_cluster_security_group;
DROP TABLE IF EXISTS cdm_cluster_instance_user_script;
DROP TABLE IF EXISTS cdm_cluster_instance;
DROP TABLE IF EXISTS cdm_cluster_keypair;
DROP TABLE IF EXISTS cdm_cluster_instance_extra_spec;
DROP TABLE IF EXISTS cdm_cluster_instance_spec;
DROP TABLE IF EXISTS cdm_cluster_volume_snapshot;
DROP TABLE IF EXISTS cdm_cluster_volume;
DROP TABLE IF EXISTS cdm_cluster_permission;
DROP TABLE IF EXISTS cdm_cluster_hypervisor;
DROP TABLE IF EXISTS cdm_cluster_availability_zone;
DROP TABLE IF EXISTS cdm_cluster_quota;
DROP TABLE IF EXISTS cdm_cluster_tenant;
DROP TABLE IF EXISTS cdm_cluster_storage;
DROP TABLE IF EXISTS cdm_cluster;

-- sequence
DROP SEQUENCE IF EXISTS cdm_cluster_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_hypervisor_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_availability_zone_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_tenant_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_network_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_floating_ip_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_subnet_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_subnet_nameserver_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_router_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_router_extra_route_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_subnet_dhcp_pool_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_security_group_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_security_group_rule_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_storage_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_volume_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_volume_snapshot_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_instance_spec_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_instance_extra_spec_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_keypair_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_instance_seq;
DROP SEQUENCE IF EXISTS cdm_cluster_instance_network_seq;

-- cdm_cluster
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_seq') ,
	tenant_id INTEGER NOT NULL,
	owner_group_id INTEGER NOT NULL,
	name VARCHAR(255) NOT NULL,
	remarks VARCHAR(300),
	type_code VARCHAR(100) NOT NULL,
	api_server_url VARCHAR(300) NOT NULL,
	credential VARCHAR(1024) NOT NULL,
	state_code VARCHAR(100) NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	synchronized_at INTEGER NOT NULL,
	PRIMARY KEY (id),
    CONSTRAINT cdm_cluster_name_un UNIQUE  (name)
);

-- cdm_cluster_permission
CREATE TABLE IF NOT EXISTS cdm_cluster_permission (
	cluster_id INTEGER NOT NULL,
	group_id INTEGER NOT NULL,
	mode_code VARCHAR(100) NOT NULL,
	PRIMARY KEY (cluster_id, group_id),
	CONSTRAINT cdm_cluster_permission_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_availability_zone
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_availability_zone_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_availability_zone (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_availability_zone_seq'),
	cluster_id INTEGER NOT NULL,
	name VARCHAR(255) NOT NULL,
	available BOOLEAN NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_availability_zone_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_hypervisor
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_hypervisor_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_hypervisor (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_hypervisor_seq'),
	cluster_id INTEGER NOT NULL,
	cluster_availability_zone_id INTEGER NOT NULL,
    uuid VARCHAR(36) NOT NULL,
    type_code VARCHAR (100) NOT NULL,
	hostname VARCHAR (255) NOT NULL,
	ip_address VARCHAR (40) NOT NULL,
	vcpu_total_cnt INTEGER NOT NULL,
	vcpu_used_cnt INTEGER NOT NULL,
	mem_total_bytes INTEGER NOT NULL,
	mem_used_bytes INTEGER NOT NULL,
	disk_total_bytes INTEGER NOT NULL,
	disk_used_bytes INTEGER NOT NULL,
	state VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	ssh_port INTEGER,
	ssh_account VARCHAR(30),
	ssh_password VARCHAR(30),
	agent_port INTEGER,
	agent_version VARCHAR (10),
	agent_installed_at INTEGER,
	agent_last_upgraded_at INTEGER,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_hypervisor_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT cdm_cluster_hypervisor_cluster_availability_zone_id_fk FOREIGN KEY (cluster_availability_zone_id) REFERENCES cdm_cluster_availability_zone (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_tenant
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_tenant_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_tenant (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_tenant_seq'),
	cluster_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description TEXT,
	enabled BOOLEAN NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_tenant_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_quota
CREATE TABLE IF NOT EXISTS cdm_cluster_quota (
	cluster_tenant_id INTEGER NOT NULL,
	key VARCHAR(255) NOT NULL,
	value INTEGER NOT NULL,
	PRIMARY KEY (cluster_tenant_id, key),
	CONSTRAINT cdm_cluster_quota_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_network
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_network_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_network (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_network_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	type_code VARCHAR(100) NOT NULL,
	external_flag BOOLEAN NOT NULL,
	state VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_network_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_floating_ip
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_floating_ip_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_floating_ip (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_floating_ip_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	cluster_network_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	description VARCHAR(255),
	ip_address VARCHAR(40) NOT NULL,
	status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_floating_ip_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_floating_ip_cluster_network_id_fk FOREIGN KEY (cluster_network_id) REFERENCES cdm_cluster_network (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_subnet
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_subnet_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_subnet (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_subnet_seq'),
	cluster_network_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	network_cidr VARCHAR(64) NOT NULL,
	dhcp_enabled BOOLEAN NOT NULL,
	gateway_enabled BOOLEAN NOT NULL,
	gateway_ip_address VARCHAR(40),
	ipv6_address_mode_code VARCHAR(100),
	ipv6_ra_mode_code VARCHAR(100),
	raw TEXT,
    PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_subnet_cluster_network_id_fk FOREIGN KEY (cluster_network_id) REFERENCES cdm_cluster_network (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_subnet_nameserver
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_subnet_nameserver_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_subnet_nameserver(
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_subnet_nameserver_seq'),
	cluster_subnet_id INTEGER NOT NULL,
	nameserver VARCHAR(255) NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_subnet_nameserver_cluster_subnet_id_fk FOREIGN KEY (cluster_subnet_id) REFERENCES cdm_cluster_subnet (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_router
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_router_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_router(
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_router_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	state VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_router_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_router_extra_route
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_router_extra_route_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_router_extra_route (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_router_extra_route_seq'),
	cluster_router_id INTEGER NOT NULL,
	destination VARCHAR(64) NOT NULL,
	nexthop VARCHAR (40) NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_router_extra_route_cluster_router_id_fk FOREIGN KEY (cluster_router_id) REFERENCES cdm_cluster_router (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_routing_interface
CREATE TABLE IF NOT EXISTS cdm_cluster_routing_interface (
	cluster_router_id INTEGER NOT NULL,
	cluster_subnet_id INTEGER NOT NULL,
	ip_address VARCHAR(40) NOT NULL,
	external_flag BOOLEAN NOT NULL,
	PRIMARY KEY (cluster_router_id, cluster_subnet_id),
	CONSTRAINT cdm_cluster_routing_interface_cluster_router_id_fk FOREIGN KEY (cluster_router_id) REFERENCES cdm_cluster_router (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_routing_interface_cluster_subnet_id_fk FOREIGN KEY (cluster_subnet_id) REFERENCES cdm_cluster_subnet (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_subnet_dhcp_pool
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_subnet_dhcp_pool_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_subnet_dhcp_pool (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_subnet_dhcp_pool_seq'),
	cluster_subnet_id INTEGER NOT NULL,
	start_ip_address VARCHAR(40) NOT NULL,
	end_ip_address VARCHAR(40) NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_subnet_dhcp_pool_cluster_networ_subnet_id_fk FOREIGN KEY (cluster_subnet_id) REFERENCES cdm_cluster_subnet (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_security_group
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_security_group_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_security_group (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_security_group_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_security_group_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_security_group_rule
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_security_group_rule_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_security_group_rule (
	id INTEGER NOT NULL DEFAULT NEXTVAL('cdm_cluster_security_group_rule_seq'),
	security_group_id INTEGER NOT NULL,
	remote_security_group_id INTEGER,
	uuid VARCHAR(36) NOT NULL,
	description VARCHAR(255),
	ether_type INTEGER NOT NULL,
	network_cidr VARCHAR(64),
	direction VARCHAR(20) NOT NULL,
	port_range_min INTEGER,
	port_range_max INTEGER,
	protocol VARCHAR(20),
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_security_group_rule_security_group_id_fk FOREIGN KEY (security_group_id) REFERENCES cdm_cluster_security_group (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_security_group_rule_remote_security_group_id_fk FOREIGN KEY (remote_security_group_id) REFERENCES cdm_cluster_security_group (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_storage
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_storage_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_storage (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_storage_seq'),
	cluster_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	type_code VARCHAR(100) NOT NULL,
	capacity_bytes INTEGER,
	used_bytes INTEGER,
	credential TEXT,
    status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_storage_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_volume
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_volume_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_volume (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_volume_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	cluster_storage_id INTEGER NOT NULL,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	size_bytes INTEGER NOT NULL,
	multiattach BOOLEAN NOT NULL,
	bootable BOOLEAN NOT NULL,
	readonly BOOLEAN NOT NULL,
	status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_volume_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_volume_cluster_storage_id_fk FOREIGN KEY (cluster_storage_id) REFERENCES cdm_cluster_storage (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_volume_snapshot
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_volume_snapshot_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_volume_snapshot (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_volume_snapshot_seq'),
	cluster_volume_id INTEGER NOT NULL,
	cluster_volume_group_snapshot_uuid VARCHAR(36),
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	size_bytes INTEGER NOT NULL,
	status VARCHAR(20) NOT NULL,
	created_at INTEGER NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_volume_snapshot_cluster_volume_id_fk FOREIGN KEY (cluster_volume_id) REFERENCES cdm_cluster_volume (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_spec
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_instance_spec_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_spec (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_instance_spec_seq'),
	cluster_id INTEGER NOT NULL,
	uuid VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	vcpu_total_cnt INTEGER NOT NULL,
	mem_total_bytes INTEGER NOT NULL,
	disk_total_bytes INTEGER NOT NULL,
	swap_total_bytes INTEGER NOT NULL,
	ephemeral_total_bytes INTEGER NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_instance_spec_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES cdm_cluster (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_extra_spec
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_instance_extra_spec_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_extra_spec (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_instance_extra_spec_seq'),
	cluster_instance_spec_id INTEGER NOT NULL,
	key VARCHAR(255) NOT NULL,
	value VARCHAR(255) NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_instancextra_spec_cluster_intance_spec_id_fk FOREIGN KEY (cluster_instance_spec_id) REFERENCES cdm_cluster_instance_spec (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_keypair
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_keypair_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_keypair (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_keypair_seq'),
	cluster_id INTEGER NOT NULL,
	name VARCHAR(255) NOT NULL,
	fingerprint VARCHAR(100) NOT NULL,
	public_key VARCHAR(2048) NOT NULL,
	type_code VARCHAR(50) NOT NULL,
	PRIMARY KEY (id)
);

-- cdm_cluster_instance
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_instance_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_instance (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_instance_seq'),
	cluster_tenant_id INTEGER NOT NULL,
	cluster_availability_zone_id INTEGER NOT NULL,
	cluster_hypervisor_id INTEGER NOT NULL,
	cluster_instance_spec_id INTEGER NOT NULL,
	cluster_keypair_id INTEGER,
	uuid VARCHAR(36) NOT NULL,
	name VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	state VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	raw TEXT,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_instance_cluster_tenant_id_fk FOREIGN KEY (cluster_tenant_id) REFERENCES cdm_cluster_tenant (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_cluster_availability_zone_id_fk FOREIGN KEY (cluster_availability_zone_id) REFERENCES cdm_cluster_availability_zone (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_cluster_hypervisor_id_fk FOREIGN KEY (cluster_hypervisor_id) REFERENCES cdm_cluster_hypervisor (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_cluster_instance_spec_id_fk FOREIGN KEY (cluster_instance_spec_id) REFERENCES cdm_cluster_instance_spec (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_cluster_keypair_id_fk FOREIGN KEY (cluster_keypair_id) REFERENCES cdm_cluster_keypair (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_network
CREATE SEQUENCE IF NOT EXISTS cdm_cluster_instance_network_seq;
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_network (
	id INTEGER NOT NULL DEFAULT NEXTVAL ('cdm_cluster_instance_network_seq'),
	cluster_instance_id INTEGER NOT NULL,
	cluster_network_id INTEGER NOT NULL,
	cluster_subnet_id INTEGER NOT NULL,
	cluster_floating_ip_id INTEGER,
	dhcp_flag BOOLEAN NOT NULL,
	ip_address VARCHAR(40) NOT NULL,
	PRIMARY KEY (id),
	CONSTRAINT cdm_cluster_instance_network_cluster_instance_id_fk FOREIGN KEY (cluster_instance_id) REFERENCES cdm_cluster_instance (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_network_cluster_network_id_fk FOREIGN KEY (cluster_network_id) REFERENCES cdm_cluster_network (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_network_cluster_subnet_fk FOREIGN KEY (cluster_subnet_id) REFERENCES cdm_cluster_subnet (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_network_cluster_floating_ip_id_fk FOREIGN KEY (cluster_floating_ip_id) REFERENCES cdm_cluster_floating_ip (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_security_group
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_security_group (
	cluster_instance_id INTEGER NOT NULL,
	cluster_security_group_id INTEGER NOT NULL,
	PRIMARY KEY (cluster_instance_id, cluster_security_group_id),
	CONSTRAINT cdm_cluster_instance_security_group_cluster_instance_id_fk FOREIGN KEY (cluster_instance_id) REFERENCES cdm_cluster_instance (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_security_group_cluster_security_group_id_fk FOREIGN KEY (cluster_security_group_id) REFERENCES cdm_cluster_security_group (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_volume 
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_volume (
	cluster_instance_id INTEGER NOT NULL,
	cluster_volume_id INTEGER NOT NULL,
	device_path VARCHAR(4096) NOT NULL,
	boot_index INTEGER NOT NULL,
	PRIMARY KEY (cluster_instance_id, cluster_volume_id),
	CONSTRAINT cdm_cluster_instance_volume_cluster_instance_id_fk FOREIGN KEY (cluster_instance_id) REFERENCES cdm_cluster_instance (id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	CONSTRAINT cdm_cluster_instance_volume_cluster_volume_id_fk FOREIGN KEY (cluster_volume_id) REFERENCES cdm_cluster_volume (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

-- cdm_cluster_instance_user_script
CREATE TABLE IF NOT EXISTS cdm_cluster_instance_user_script (
    cluster_instance_id INTEGER NOT NULL,
    user_data TEXT NOT NULL,
    PRIMARY KEY (cluster_instance_id),
    CONSTRAINT cdm_cluster_instance_script_cluster_instance_id_fk FOREIGN KEY (cluster_instance_id) REFERENCES cdm_cluster_instance (id) ON UPDATE RESTRICT ON DELETE RESTRICT
);
