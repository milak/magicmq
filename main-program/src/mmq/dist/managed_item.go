package dist

import (
	"mmq/item"
)

type ManagedItem struct {
	Item 		*item.Item
	Instances	[]string
}
func (this *ManagedItem) AddInstance(aInstance string) {
	this.Instances = append(this.Instances,aInstance)
}