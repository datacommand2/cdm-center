package openstack

// InstanceStateMap 인스턴스 state 값
var InstanceStateMap = map[int64]string{
	0: "NOSTATE",
	1: "RUNNING",
	3: "PAUSED",
	4: "SHUTDOWN",
	6: "CRASHED",
	7: "SUSPENDED",
}
