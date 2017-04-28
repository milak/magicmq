package service

import (
	"mmq/env"
	"mmq/item"
	"mmq/conf"
	"mmq/dist"
	"github.com/milak/event"
	"github.com/milak/math"
	"strconv"
	"strings"
)

type ItemProcessorService struct {
	context 	*env.Context
	pool 		*dist.InstancePool
}

func NewItemProcessorService(aContext *env.Context,aPool *dist.InstancePool) *ItemProcessorService {
	result := &ItemProcessorService{context: aContext,pool:aPool}
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
			var nbInstances = 0
			for _, i := range this.context.Configuration.Instances {
				if i.Connected {
					nbInstances++
				}
			}
			if nbInstances == 0 {
				this.context.Logger.Println("Unable to apply distribution : no connected instances")
				return
			}
			this.context.Logger.Println("I am connected with :",nbInstances,"instances")
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
			e.Item.SetShared(true)
			var sharedItem dist.SharedItem
			sharedItem.Item = e.Item
			for _, i := range this.context.Configuration.Instances {
				if count <= 0 {
					break
				}
				if i.Connected {
					this.context.Logger.Println("Selected",i)
					sharedItem.AddInstance(i.Name())
					count--
				}
			}
			for _,i := range sharedItem.Instances {
				instanceConnection := this.pool.GetInstanceByName(i)
				if instanceConnection != nil {
					this.context.Logger.Println("Distributing with",i)
					instanceConnection.SendItem(sharedItem)
					writer := instanceConnection.SendItemContent(e.Item.ID,e.Item.Size())
					bytes := make([]byte,2000)
					count,_ := e.Item.Read(bytes)
					for count != -1 {
						writer.Write(bytes[0:count])
						count,_ = e.Item.Read(bytes)
					}
				} else {
					this.context.Logger.Println("WARNING unable to get connection to ",i)
				}
			}
		}
	}
}
func (this *ItemProcessorService) Start() {
	event.EventBus.AddListener(this)
}
func (this *ItemProcessorService) Stop() {
	event.EventBus.RemoveListener(this)
}
