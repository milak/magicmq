package item

import (
	"mmq/env"
	"mmq/conf"
	"github.com/milak/event"
	"fmt"
	"errors"
)

type ItemStore struct {
	itemsByTopic map[string][]*Item
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
	result := &ItemStore{itemsByTopic : make(map[string][]*Item), context : aContext}
	return result
}
func (this *ItemStore) Push (aItem *Item) error {
	// Pour chaque topic pour lequel il est enregistré
	for _,topicName := range aItem.Topics {
		topic := this.context.Configuration.GetTopic(topicName)
		if topic == nil {
			return StoreError{"Topic not found",topicName,"nil"}
		}
		if topic.Type == conf.VIRTUAL {
			return StoreError{"Unable to push in virtual topic",topicName,"nil"}
		}
		items := this.itemsByTopic[topicName]
		if items == nil {
			this.itemsByTopic[topicName] = make([]*Item,1)
			this.itemsByTopic[topicName][0] = aItem
		} else {
			this.itemsByTopic[topicName] = append(this.itemsByTopic[topicName],aItem)
		}
		event.EventBus.FireEvent(&ItemAdded{aItem,topic})
	}
	return nil
}
func (this *ItemStore) Pop(aTopicName string) (*Item, error) {
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
		strategy := topic.GetParameterByName(conf.PARAMETER_STRATEGY)
		if strategy == "" {
			strategy = conf.ORDERED
		}
		if strategy == conf.ORDERED {
			// Par défaut on est en mode ORDERED : on vide le premier topic avant de vider le second
			for _,subTopicName := range subTopics {
				//faut-il vérifier que le topic existe ? subTopic := this.configuration.GetTopic(subTopicName)
				items := this.itemsByTopic[subTopicName]
				if len(items) != 0 {
					item := items[0]
					this.itemsByTopic[subTopicName] = items[1:]
					return item, nil
				}
			}
		} else if strategy == conf.ROUND_ROBIN {
			// TODO implémenter la stratégie ROUND-ROBIN 
			// Pour implémenter ROUND-ROBIN, il va falloir conserver un indicateur pour savoir la file que l'on a lu le coup précédent
			return nil, errors.New("ROUND ROBIN strategy not yet implemented")
		} else {
			return nil, errors.New(strategy + " strategy not recognized")
		}
		return nil, nil
	}
}
func (this *ItemStore) RemoveItem(aTopicName string, aItem *Item) {
	items := this.itemsByTopic[aTopicName]
	for i,item := range items {
		if item.ID == aItem.ID {
			this.itemsByTopic[aTopicName] = append(items[0:i],items[i+1:]...)
			return
		}
	}
}
func (this *ItemStore) List(aTopicName string) ([]*Item, error) {
	topic := this.context.Configuration.GetTopic(aTopicName)
	if topic == nil {
		return nil, StoreError{"Topic not found",aTopicName,"nil"}
	}
	var result []*Item = nil
	if topic.Type == conf.SIMPLE {
		result = this.itemsByTopic[aTopicName]
		if result == nil {
			topic := this.context.Configuration.GetTopic(aTopicName)
			if topic != nil {
				result = []*Item{}
			}
		}
	} else {
		subTopics := topic.TopicList
		strategy := topic.GetParameterByName(conf.PARAMETER_STRATEGY)
		if strategy == "" {
			strategy = conf.ORDERED
		}
		if strategy == conf.ORDERED {
			result = []*Item{}
			// Par défaut on est en mode ORDERED : on vide le premier topic avant de vider le second
			for _,subTopicName := range subTopics {
				//faut-il vérifier que le topic existe ? subTopic := this.configuration.GetTopic(subTopicName)
				items := this.itemsByTopic[subTopicName]
				for _,item := range items {
					result = append(result,item)
				}
			}
		} else if strategy == conf.ROUND_ROBIN {
			result = []*Item{}
			// TODO implémenter la stratégie ROUND-ROBIN 
			// on alterne dans chaque sub topic
			topics := []string{}
			for _,subTopicName := range subTopics {
				topics = append(topics,subTopicName)
			}
			for len(topics) > 0 {
				newTopics := []string{}
				for _,subTopicName := range topics {
					items := this.itemsByTopic[subTopicName]
					if len(items) == 0 {
						continue
					}
					result = append(result,items[0])
					newTopics = append(newTopics,subTopicName)
				}
				topics = newTopics
			}
		} else {
			return nil, errors.New(strategy + " strategy not recognized")
		}
	}
	return result, nil
}
func (this *ItemStore) Count(aTopicName string) int {
	return len(this.itemsByTopic[aTopicName])
}