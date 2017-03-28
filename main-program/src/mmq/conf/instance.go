package conf

import (

)

type Instance struct {
	Host string
	Port string
}
func NewInstance(aHost string, aPort string) *Instance{
	return &Instance {Host : aHost, Port: aPort}
}
func (this *Instance) Connect() {
	
}
