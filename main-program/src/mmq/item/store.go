package item

import (
	
)
type ItemStore struct {
	itemsByTopic map[string][]Item
}
func NewStore() *ItemStore{
	return &ItemStore{itemsByTopic : make(map[string][]Item)}
}
func (this *ItemStore)Push(aItem Item){
	topics := aItem.Topics()
	// Pour chaque topic pour lequel il est enregistré
	for t := range topics {
		items := this.itemsByTopic[topics[t]]
		if items == nil {
			this.itemsByTopic[topics[t]] = make([]Item,1)
			this.itemsByTopic[topics[t]][0] = aItem
		} else {
			this.itemsByTopic[topics[t]] = append(this.itemsByTopic[topics[t]],aItem)
		}
	}
}
func (this *ItemStore)Pop(aTopic string) Item{
	items := this.itemsByTopic[aTopic]
	if len(items) == 0 {
		return nil
	} else {
		item := items[0]
		this.itemsByTopic[aTopic] = items[1:]
		return item
	}
}
func (this *ItemStore)Count(aTopic string) int{
	return len(this.itemsByTopic[aTopic])
}