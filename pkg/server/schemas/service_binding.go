package schemas

type ServiceBindingRequest struct {
	Context      map[string]string           `json:"context,omitempty"`
	ServiceId    string                      `json:"service_id"`
	PlanId       string                      `json:"plan_id"`
	BindResource ServiceBindingResouceObject `json:"bind_resource,omitempty"`
	Parameters   map[string]string           `json:"parameters,omitempty"`
}

type ServiceBindingResouceObject struct {
	AppGuid string `json:"app_guid,omitempty"`
	Route   string `json:"route,omitempty"`
}

type ServiceBindingResponse struct {
	Metadata        ServiceBindingMetadata    `json:"metadata,omitempty"`
	Credentials     map[string]string         `json:"credentials,omitempty"`
	SyslogDrainUrl  string                    `json:"syslog_drain_url,omitempty"`
	RouteServiceUrl string                    `json:"route_service_url,omitempty"`
	VolumeMounts    ServiceBindingVolumeMount `json:"volume_mounts,omitempty"`
	Endpoints       ServiceBindingEndpoint    `json:"endpoints,omitempty"`
}

type ServiceBindingMetadata struct {
	ExpiresAt string `json:"expires_at,omitemtpy"`
}

type ServiceBindingVolumeMount struct {
	Driver          string                          `json:"driver"`
	ContainerDriver string                          `json:"container_driver"`
	Mode            string                          `json:"mode"`
	DeviceType      string                          `json:"device_type"`
	Device          ServiceBindingVolumeMountDevice `json:"device"`
}

type ServiceBindingVolumeMountDevice struct {
	VolumeId    string            `json:"volume_id"`
	MountConfig map[string]string `json:"mount_config,omitempty"`
}

type ServiceBindingEndpoint struct {
	Host     string   `json:"host"`
	Ports    []string `json:"ports"`
	Protocol string   `json:"protocol,omitempty"`
}

type AsyncOperation struct {
	Operation string `json:"operation"`
}
