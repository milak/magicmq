package conf

import (

)

type Instance struct {
	Name string
	host string
	port string
	Connected bool
}
func NewInstance(aHost string, aPort string) *Instance{
	return &Instance {Name : aHost+":"+aPort, host : aHost, port: aPort, Connected : false}
}
func (this *Instance) Connect() {
	
}
