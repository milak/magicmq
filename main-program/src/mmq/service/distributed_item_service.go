package service

import (
	"mmq/env"
	"mmq/item"
	"mmq/conf"
	"github.com/milak/event"
	"github.com/milak/math"
	"strconv"
	"strings"
)

type ItemProcessorService struct {
	context *env.Context
}

func NewItemProcessorService(aContext *env.Context) *ItemProcessorService {
	result := &ItemProcessorService{context: aContext}
	return result
}

// Catch event ItemAdded
func (this *ItemProcessorService) Event(aEvent interface{}) {
	switch e := aEvent.(type) {
	case item.ItemAdded:
		if e.Topic.IsDistributed() {
			distributionPolicy := *(e.Topic.GetParameterByName(conf.PARAMETER_DISTRIBUTED))
			var nbInstances = 0
			for _, i := range this.context.Configuration.Instances {
				if i.Connected {
					nbInstances++
				}
			}
			nbInstances++ // on ajoute l'instance courante en plus
			var count int
			var err error
			// ne devrait pas être == conf.DISTRIBUTED_NO
			if distributionPolicy == conf.DISTRIBUTED_ALL {
				count = nbInstances
			} else if strings.HasSuffix(distributionPolicy, "%") {
				percent, err := strconv.Atoi(distributionPolicy[0 : len(distributionPolicy)-1])
				if err != nil {
					this.context.Logger.Println("Unable to apply distribution : ", distributionPolicy, e.Topic.Name, "nil")
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
			// distribute item to other instances
			for count > 1 {

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
