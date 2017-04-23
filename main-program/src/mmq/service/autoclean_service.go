package service

import (
	"mmq/conf"
	"mmq/env"
	"mmq/item"
	"time"
)
type AutoCleanService struct {
	context *env.Context
	store   *item.ItemStore
	running bool
}
func NewAutoCleanService(aContext *env.Context, aStore *item.ItemStore) *AutoCleanService {
	result := &AutoCleanService{context:aContext, store : aStore, running : false}
	return result;
}
func (this *AutoCleanService) Start(){
	if !this.running {
		this.running = true
		go this.run()
	}
}
func (this *AutoCleanService) Stop(){
	this.running = false
}
func (this *AutoCleanService) run (){
	this.context.Logger.Println("Starting autoclean")
	// TODO prendre en compte lorsqu'un nouveau TOPIC est ajouté ou mis à jour via les évènements
	topics, timeToLives := computeTimeToLivesAndTopics(this.context)
	for this.running && this.context.Running {
		time.Sleep(1 * time.Second)
		//this.context.Logger.Println("Cleaning")
		for topicIndex,topic := range topics {
			//this.context.Logger.Println("Topic ",topic.Name," ",timeToLives[topicIndex])
			for _,item := range this.store.List(topic.Name) {
				age := item.GetAge()
				this.context.Logger.Println("Computing for ",item," ",age)
				if age > timeToLives[topicIndex] {
					//fmt.Println("Removing ",item)
					this.store.RemoveItem(topic.Name,item)
				}
			}
			
		}
	}
}
func computeTimeToLivesAndTopics(aContext *env.Context) ([]*conf.Topic, []time.Duration) {
	var topics []*conf.Topic
	var timeToLives []time.Duration
	for _,topic := range aContext.Configuration.Topics {
		duration,err := topic.GetTimeToLive()
		if err != nil {
			aContext.Logger.Println("WARNING Unable to parse PARAMETER " + conf.PARAMETER_TIME_TO_LIVE + " '"+topic.GetParameterByName(conf.PARAMETER_TIME_TO_LIVE)+"' for topic " +topic.Name + " time to live will not be used",err)
			continue
		}
		if duration == nil {
			continue
		}
		timeToLives = append(timeToLives,*duration)
		topics = append(topics, topic)
	}
	return topics, timeToLives
}
