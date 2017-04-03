package conf

import (
	"os"
	"encoding/json"
	"fmt"
)
const APP_VERSION = "0.1" // The version of the current application
type Configuration struct {
	Version 	string
	Topics 		[]*Topic 		`json:"Topics,omitempty"`
	Instances 	[]*Instance 	`json:"Instances,omitempty"`
	fileName 	string
	Services 	[]Service
}
type Service struct {
	Name 		string
	Comment 	string
	Active 		bool
	Parameters 	[]Parameter `json:"Parameters,omitempty"`
}
type Parameter struct {
	Name 		string
	Value 		string
}
func (this *Configuration) AddInstance(aInstance *Instance) bool{
	if this.GetInstance(aInstance.Name()) != nil {
		return false
	}
	fmt.Println("configuration.go: Adding instance",aInstance)
	this.Instances = append(this.Instances,aInstance)
	this.save()
	return true
}
func (this *Configuration) GetInstance(aName string) *Instance{
	for _,instance := range this.Instances {
		if instance.Name() == aName {
			return instance
		}
	}
	return nil
}
/**
 * Remove a instance by name
 */
func (this *Configuration) RemoveInstance(aInstanceName string) *Instance {
	found := false
	fmt.Println("RemoveInstance",aInstanceName)
	var instance *Instance
	for i := range this.Instances {
		instance = this.Instances[i]
		if (instance.Name() == aInstanceName){
			this.Instances = append(this.Instances[0:i],this.Instances[i+1:]...)
			found = true
			break;
		}
	}
	if found {
		this.save()
	}
	return instance
}
/**
 * Add a topic in the list
 */
func (this *Configuration) AddTopic(aTopic *Topic) bool {
	if this.GetTopic(aTopic.Name) != nil {
		return false
	}
	this.Topics = append(this.Topics,aTopic)
	this.save()
	return true
}
/**
 * Get a topic by name
 */
func (this *Configuration) GetTopic(aName string) *Topic {
	for _,topic := range this.Topics {
		if topic.Name == aName {
			return topic
		}
	}
	return nil
}
/**
 * Remove a topic by name
 */
func (this *Configuration) RemoveTopic(aTopicName string) bool {
	found := false
	for i := range this.Topics {
		topic := this.Topics[i]
		if (topic.Name == aTopicName){
			this.Topics = append(this.Topics[0:i],this.Topics[i+1:]...)
			found = true
			break;
		}
	}
	if found {
		this.save()
	}
	return found
}
/**
 * Swap the configuration in file
 */
func (this *Configuration) save(){
	file,err := os.Create(this.fileName)
	if err != nil {
		panic ("Unable to write file " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	encoder.Encode(this)
}
/**
 * Constructor of Configuration if the file exists, the configuration is loaded, if the file doesn't exist, the file is created with default configuration
 */
func InitConfiguration(aFileName string) *Configuration {
	result := Configuration{Version 		: APP_VERSION,fileName 		: aFileName}
	if _, err := os.Stat(aFileName); os.IsNotExist(err) {
		result.Services = make([]Service,3)
		result.Services[0].Name = "ADMIN"
		result.Services[0].Comment = "This service opens web administration. It requires REST service. Parameter : 'root' directory containing admin web files. Can be replaced by apache httpd."
		result.Services[0].Active = true
		result.Services[0].Parameters = make([]Parameter,1)
		result.Services[0].Parameters[0].Name = "root"
		result.Services[0].Parameters[0].Value = "web"
		result.Services[1].Name = "REST"
		result.Services[1].Comment = "This service opens REST API. Parameter : 'port' the listening port."
		result.Services[1].Active = true
		result.Services[1].Parameters = make([]Parameter,1)
		result.Services[1].Parameters[0].Name = "port"
		result.Services[1].Parameters[0].Value = "8080"
		result.Services[2].Name = "SYNC"
		result.Services[2].Comment = "This service opens SYNC port for clusterisation. Parameter : 'port' the listening port."
		result.Services[2].Active = true
		result.Services[2].Parameters = make([]Parameter,1)
		result.Services[2].Parameters[0].Name = "port"
		result.Services[2].Parameters[0].Value = "8080"
		/*result.Services[3].Name = "PROTOBUF"
		result.Services[3].Comment = "TODO service"
		result.Services[3].Active = false*/
		result.save()
	} else {
		file,_ := os.Open(aFileName)
		defer file.Close()
		decoder := json.NewDecoder(file)
		decoder.Decode(&result)
		if result.Topics == nil {
			result.Topics = make ([]*Topic,0)
		}
		if result.Instances == nil {
			result.Instances = make ([]*Instance,0)
		}
		// Initialiser convenablement les instances à "non connectées"
		for _,instance := range result.Instances {
			instance.Connected = false
		}
	}
	return &result
}