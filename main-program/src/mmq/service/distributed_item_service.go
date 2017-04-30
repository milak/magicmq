package service

import (
	"bytes"
	"mmq/env"
	"mmq/item"
	"mmq/conf"
	"mmq/dist"
	"github.com/milak/event"
	"github.com/milak/math"
	"log"
	"math/rand"
	"time"
	"strconv"
	"strings"
)

type ItemProcessorService struct {
	context 				*env.Context
	pool 					*dist.InstancePool
	store					*item.ItemStore
	logger					*log.Logger
	sharedItems 			map[string]*dist.SharedItem
	contents				map[string][]byte
	receivedItemsByInstance	map[string][]*dist.SharedItem
}

func NewItemProcessorService(aContext *env.Context,aPool *dist.InstancePool, aStore *item.ItemStore) *ItemProcessorService {
	result := &ItemProcessorService{context: aContext,pool:aPool, store : aStore}
	result.sharedItems = make (map[string]*dist.SharedItem)
	result.receivedItemsByInstance = make(map[string][]*dist.SharedItem)
	result.contents = make(map[string][]byte)
	result.logger = aContext.Logger
	return result
}

// Catch event ItemAdded
func (this *ItemProcessorService) Event(aEvent interface{}) {
	switch e := aEvent.(type) {
	case *item.ItemAdded:
		if e.Topic.IsDistributed() {
			distributionPolicy := e.Topic.GetParameterByName(conf.PARAMETER_DISTRIBUTED)
			this.context.Logger.Println("Received new distributed item policy :",distributionPolicy)
			if distributionPolicy == "0" || distributionPolicy == "1" || distributionPolicy == "" {
				this.context.Logger.Println("No need to distribute : distributionPolicy == '"+distributionPolicy+"'")
				return
			}
			groups := e.Topic.GetParameterByName(conf.PARAMETER_DISTRIBUTED_GROUPS)
			if groups == "" {
				groups = "all"
			}
			var nbInstances = 0
			var retainedInstances []*conf.Instance
			for _, i := range this.context.Configuration.Instances {
				if i.Connected {
					found := false
					for _,g := range i.Groups {
						if g == groups {
							found = true
							break
						}
					}
					if found  {
						retainedInstances = append(retainedInstances,i)
						nbInstances++
					}
				}
			}
			if nbInstances == 0 {
				this.context.Logger.Println("Unable to apply distribution : no connected instances")
				return
			}
			this.context.Logger.Println("I am connected with :",nbInstances,"instances",retainedInstances)
			nbInstances++ // on ajoute l'instance courante en plus
			var count int
			var err error
			// ne devrait pas être == conf.DISTRIBUTED_NO
			if distributionPolicy == conf.DISTRIBUTED_ALL {
				count = nbInstances
			} else if strings.HasSuffix(distributionPolicy, "%") {
				percent, err := strconv.Atoi(distributionPolicy[0 : len(distributionPolicy)-1])
				if err != nil {
					this.context.Logger.Println("Unable to apply distribution : ", distributionPolicy, e.Topic.Name, "nil", err)
					return
				}
				count = (percent * nbInstances / 100)
				if math.Odd(nbInstances) {
					count++
				}
				if count == 0 {
					count = 1
				}
			} else {
				count, err = strconv.Atoi(distributionPolicy)
				if err != nil {
					this.context.Logger.Println("Unable to apply distribution : ", distributionPolicy, e.Topic.Name, "nil")
					return
				}
				if count > nbInstances {
					count = nbInstances
				}
			}
			this.context.Logger.Println("The item must be distributed with :",count,"instances (including me)")
			count-- // removing myself
			this.context.Logger.Println("Looking for :",count," instances to share with me")
			// distribute item to other instances
			now := time.Now()
			rand.Seed(int64(now.Nanosecond())) // Try changing this number!
			ids := rand.Perm(nbInstances-1)
			this.context.Logger.Println("IDS ",ids)
			e.Item.SetShared(true)
			var sharedItem dist.SharedItem
			sharedItem.Topic = e.Topic.Name
			this.sharedItems[e.Item.ID] = &sharedItem
			sharedItem.Item = e.Item
			for count > 0 {
				id := ids[count-1]
				i := retainedInstances[id]
				this.context.Logger.Println("Selected",i)
				sharedItem.AddInstance(i.Name())
				count--
			}
			for _,i := range sharedItem.Instances {
				instanceConnection := this.pool.GetInstanceByName(i)
				if instanceConnection != nil {
					this.context.Logger.Println("Distributing with",i)
					instanceConnection.SendItem(sharedItem)
					bytes := make([]byte,2000)
					reader := this.store.GetContent(e.Item.ID,false)
					count,_ := reader.Read(bytes)
					writer := instanceConnection.SendItemContent(e.Item.ID,count)
					
					writer.Write(bytes[0:count])
					/** TODO support BIG FILE
					for count >= 0 {
						writer.Write(bytes[0:count])
						count,_ = e.Item.Read(bytes)
					}*/
				} else {
					this.context.Logger.Println("WARNING unable to get connection to ",i)
				}
			}
		}
	case *item.ItemRemoved:
		// When a shared item is removed
		if !e.Item.IsShared() {
			return
		}
		this.context.Logger.Println("DEBUG shared item removed",e.Item)
		sharedItem := this.sharedItems[e.Item.ID]
		if sharedItem == nil {
			// TODO see whether it is normal, what to do ?
			return
		}
		delete(this.sharedItems,e.Item.ID)
		delete(this.contents,e.Item.ID)
		for _,i := range sharedItem.Instances {
			instanceConnection := this.pool.GetInstanceByName(i)
			if instanceConnection != nil {
				this.context.Logger.Println("Saying it to",i)
				instanceConnection.SendRemoveItem(e.Item.ID)
			} else {
				this.context.Logger.Println("WARNING unable to get connection to ",i)
			}
		}
	case *dist.ItemReceived :
		this.logger.Println("ItemReceived item :",e.Item," from :",e.From)
		sharedItems := this.receivedItemsByInstance[e.From.Name()]
		this.receivedItemsByInstance[e.From.Name()] = append(sharedItems,e.Item)
		this.logger.Println("- now contains :")
		for _,i := range this.receivedItemsByInstance[e.From.Name()] {
			this.logger.Println(i)
		}
	case *dist.ItemContentReceived :
		this.logger.Println("ItemContentReceived item :",e.ID," from :",e.From)
		this.logger.Println("ItemContentReceived content of the item",string(e.Content))
		this.contents[e.ID] = e.Content
	case *dist.ItemRemoved :
		this.logger.Println("ItemRemove item :",e.ID," from :",e.From)
		sharedItems := this.receivedItemsByInstance[e.From.Name()]
		for i,item := range sharedItems {
			if item.Item.ID == e.ID {
				this.receivedItemsByInstance[e.From.Name()] = append(sharedItems[0:i],sharedItems[i+1:]...)
				break
			}
		}
		this.logger.Println("- now contains :")
		for _,i := range this.receivedItemsByInstance[e.From.Name()] {
			this.logger.Println(i)
		}
	case *dist.InstanceDisconnected :
		me := this.context.InstanceName
		this.logger.Println("Instance disconnected ",e.Instance.Name())
		sharedItems := this.receivedItemsByInstance[e.Instance.Name()]
		if sharedItems != nil && len(sharedItems) != 0 {
			for _,sharedItem := range sharedItems {
				topic := this.context.Configuration.GetTopic(sharedItem.Topic)
				if topic == nil {
					this.logger.Println("WARNING a shared item is stored in unknown topic",sharedItem.Topic)
					continue
				}
				// The master instance is KO, i got item shared, let's see if i am the next MASTER
				if len(sharedItem.Instances) == 0{
					this.logger.Println("WARNING a shared instance is not shared to any one ???")
				} else {
					if sharedItem.Instances[0] != me {
						// i am not the next
						// I keep the content of the item
						continue
					}
					strategy := topic.GetParameterByName(conf.PARAMETER_DISTRIBUTION_STRATEGY)
					if strategy == "" {
						strategy = conf.PARAMETER_DISTRIBUTION_STRATEGY_AT_LEAST_ONCE
					}
					/**
					TODO take account the strategy
					PARAMETER_DISTRIBUTION_STRATEGY_AT_LEAST_ONCE
					PARAMETER_DISTRIBUTION_STRATEGY_EXACTLY_ONCE
					PARAMETER_DISTRIBUTION_STRATEGY_AT_MOST_ONCE
					*/
					// I have to take this item as mine
					// Add this item as if it would have been received for me
					// Ensure this item will be stored in only this topic
					sharedItem.Item.Topics = []string{topic.Name}
					data := this.contents[sharedItem.Item.ID]
					this.logger.Println("DEBUG content of the item to reuse ",string(data))
					buffer := bytes.NewBuffer(data)
					this.store.Push(sharedItem.Item,buffer)
					delete(this.contents,sharedItem.Item.ID) // no need to keep because i own it
				}
			}
		}
		delete(this.receivedItemsByInstance,e.Instance.Name())
	}
}
func (this *ItemProcessorService) Start() {
	event.EventBus.AddListener(this)
}
func (this *ItemProcessorService) Stop() {
	event.EventBus.RemoveListener(this)
}
