package service

import (
	"log"
	"mmq/conf"
	"mmq/env"
	"mmq/dist"
	"github.com/milak/event"
	"reflect"
	"time"
)
/**
 * The source part is about synchronisation between instances
 */


/**
 * The main class of the source code
 */
type SyncService struct {
	running 	bool							// boolean indicating if the service is running, setting it to false, should stop listening 
	context 	*env.Context					// a reference to the context, usefull to get accès to store, logger and configuration
	logger		*log.Logger						// the logger obtained from context, it is copied here for code readability reason
	pool		*dist.InstancePool
	port		string
}
/**
 * Constructor for the SyncService class
 */
func NewSyncService (aContext *env.Context, aInstancePool *dist.InstancePool) *SyncService {
	result := &SyncService{running : true, context : aContext, logger : aContext.Logger, pool : aInstancePool}
	event.EventBus.AddListener(result)
	return result
}
func (this *SyncService) Start (){
	for s := range this.context.Configuration.Services {
		service := this.context.Configuration.Services[s]
		if !service.Active {
			continue
		}
		if service.Name == "SYNC" {
			for _,p := range service.Parameters {
				if p.Name == "port" {
					this.port = p.Value
				}
			}
			this.logger.Println("starting...")
			go this.scanInstances()
			break
		}
	}
}
// Catch event InstanceRemoved
func (this *SyncService) Event(aEvent interface{}) {
	this.logger.Println("Event received")
	switch e:= aEvent.(type) {
		case *conf.InstanceRemoved :
			instanceConnection := this.pool.GetInstanceByName(e.Instance.Name())
			if instanceConnection != nil {
				instanceConnection.Close()
			}
		case *dist.TopicReceived :
			this.logger.Println("TopicReceived")
			this.logger.Println("Received topic : " + e.Topic.Name)
			existingTopic := this.context.Configuration.GetTopic(e.Topic.Name)
			if existingTopic != nil {
				this.logger.Println("Skipped because allready known")
			} else {
				this.context.Configuration.AddTopic(e.Topic)
			}
		case *dist.InstanceReceived :
			this.logger.Println("InstanceReceived")
			if (e.Instance.Host == this.context.Host) && (e.Instance.Port == this.port) {
				this.logger.Println("Skipped instance cause it is me :)")
			} else {
				e.Instance.Connected = false // ensure the Instance will not be considered as connected
				if this.context.Configuration.AddInstance(e.Instance) {
					this.logger.Println("Added instance :",e.Instance)
				}
			}
		case *dist.InstanceDisconnected :
			this.logger.Println("InstanceDisconnected")
			instanceConnection := this.pool.GetInstanceByName(e.Instance.Name())
			if instanceConnection != nil {
				instanceConnection.Close()
			}
		case *dist.ItemReceived :
			this.logger.Println("ItemReceived item :",e.Item," from :",e.From)
			
		default:
			this.logger.Println("Unknown",reflect.TypeOf(aEvent))
	}
}
/**
 * Scan not connected Instances and try to Connect
 */
func (this *SyncService) scanInstances() {
	time.Sleep(2 * time.Second)
	for this.running {
		for _,instance := range this.context.Configuration.Instances {
			if !instance.Connected {
				err := this.pool.Connect(instance)
				if err != nil {
					this.logger.Println("Error while connection to ",instance.Name(),err.Error())
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (this *SyncService) Stop (){
	this.running = false
}