package env

import (
	"mmq/conf"
	"mmq/item"
	"log"
	"os"
)

type Context struct {
	Running			bool
	Store 			*item.ItemStore
	Configuration 	*conf.Configuration
	Logger			*log.Logger
	listeners		[]ContextListener
	Host			string // will be obtained once connected, not sure it is operationnal
}
type ContextListener interface {
	TopicAdded (aTopic *conf.Topic)
	InstanceRemoved (aInstance *conf.Instance)
}
func NewContext() *Context {
	var logger *log.Logger
	file, err := os.Create("mmq.log")
	if err != nil {
		logger = log.New(os.Stdout, "-", log.Lshortfile)
		logger.Println("Unable to open file ")
	} else {
		logger = log.New(file, "-", log.Lshortfile)
	}
	return &Context{Running : true, Logger : logger}
}
func (this *Context) AddContextListener(aListener ContextListener) {
	this.listeners = append(this.listeners,aListener)
}
func (this *Context) FireTopicAdded(aTopic *conf.Topic) {
	for _,listener := range this.listeners {
		listener.TopicAdded(aTopic)
	}
}
func (this *Context) FireInstanceRemoved(aInstance *conf.Instance) {
	for _,listener := range this.listeners {
		listener.InstanceRemoved(aInstance)
	}
}