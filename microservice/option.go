package microservice

import (
	"time"
)

type ServiceOption struct {
	basePath  string
	name      string
	mode      string
	metadata  map[string]interface{}
	host      string
	port      int
	heartbeat time.Duration
}

func (o *ServiceOption) Port() int {
	return o.port
}

func (o *ServiceOption) SetPort(port int) {
	o.port = port
}

func (o *ServiceOption) Host() string {
	return o.host
}

func (o *ServiceOption) SetHost(host string) {
	o.host = host
}

func (o *ServiceOption) Metadata() map[string]interface{} {
	return o.metadata
}

func (o *ServiceOption) SetMetadata(metadata map[string]interface{}) {
	o.metadata = metadata
}

func (o *ServiceOption) Mode() string {
	return o.mode
}

func (o *ServiceOption) SetMode(mode string) {
	o.mode = mode
}

func (o *ServiceOption) Name() string {
	return o.name
}

func (o *ServiceOption) SetName(name string) {
	o.name = name
}

func (o *ServiceOption) Heartbeat() time.Duration {
	return o.heartbeat
}

func (o *ServiceOption) SetHeartbeat(heartbeat time.Duration) {
	o.heartbeat = heartbeat
}

func (o *ServiceOption) BasePath() string {
	return o.basePath
}

func (o *ServiceOption) SetBasePath(basePath string) {
	o.basePath = basePath
}
