package types

import (

)

type Configuration struct {
	APP_VERSION string
	WebDirectory string
	WebAdminPort string
	Topics []*Topic
	Instances InstanceList
}
func InitConfiguration() Configuration {
	return Configuration{APP_VERSION : "0.1", WebDirectory : "web", WebAdminPort : "8080"}
}