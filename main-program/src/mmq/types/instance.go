package types

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
type InstanceList struct {
	List []*Instance
}
func (this *InstanceList) Add(aInstance *Instance){
	this.List = append(this.List,aInstance)
}