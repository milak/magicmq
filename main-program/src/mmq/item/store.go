package item

import (
	"mmq/env"
	"mmq/conf"
	"github.com/milak/event"
	"fmt"
)

type ItemStore struct {
	itemsByTopic map[string][]Item
	context *env.Context
}

type StoreError struct {
	Message string
	Topic 	string
	Item 	string
}
func (this StoreError) Error() string {
	return fmt.Sprintf("%s : topic = %s item = %s", this.Message, this.Topic, this.Item)
}
func NewStore(aContext *env.Context) *ItemStore{
	return &ItemStore{itemsByTopic : make(map[string][]Item), context : aContext}
}
func (this *ItemStore) Push (aItem Item) error {
	// Pour chaque topic pour lequel il est enregistré
	for _,topicName := range aItem.Topics() {
		topic := this.context.Configuration.GetTopic(topicName)
		if topic == nil {
			return StoreError{"Topic not found",topicName,"nil"}
		}
		if topic.Type == conf.VIRTUAL {
			return StoreError{"Unable to push in virtual topic",topicName,"nil"}
		}
		items := this.itemsByTopic[topicName]
		if items == nil {
			this.itemsByTopic[topicName] = make([]Item,1)
			this.itemsByTopic[topicName][0] = aItem
		} else {
			this.itemsByTopic[topicName] = append(this.itemsByTopic[topicName],aItem)
		}
		event.EventBus.FireEvent(&ItemAdded{aItem,topic})
	}
	return nil
}
func (this *ItemStore) Pop(aTopicName string) (Item, error) {
	topic := this.context.Configuration.GetTopic(aTopicName)
	if topic == nil {
		return nil, StoreError{"Topic not found",aTopicName,"nil"}
	}
	if topic.Type == conf.SIMPLE {
		items := this.itemsByTopic[aTopicName]
		if len(items) == 0 {
			return nil, nil
		} else {
			item := items[0]
			this.itemsByTopic[aTopicName] = items[1:]
			return item, nil
		}
	} else {
		subTopics := topic.TopicList
		// TODO prendre en compte la stratégie : 
		// strategy := topic.GetParameterByName(conf.STRATEGY)
		// Par défaut on est en mode ORDERED : on vide le premier topic avant de vider le second
		// Pour implémenter ROUND-ROBIN, il va falloir conserver un indicateur pour savoir la file que l'on a lu le coup précédent  
		for _,subTopicName := range subTopics {
			//faut-il vérifier que le topic existe ? subTopic := this.configuration.GetTopic(subTopicName)
			items := this.itemsByTopic[subTopicName]
			if len(items) != 0 {
				item := items[0]
				this.itemsByTopic[subTopicName] = items[1:]
				return item, nil
			}
		}
		return nil, nil
	}
}
func (this *ItemStore)Count(aTopicName string) int {
	return len(this.itemsByTopic[aTopicName])
}