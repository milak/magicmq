package dist

import (
	"mmq/item"
)

type SharedItem struct {
	Item 		*item.Item
	Instances	[]string
}
func (this *SharedItem) AddInstance(aInstance string) {
	this.Instances = append(this.Instances,aInstance)
}