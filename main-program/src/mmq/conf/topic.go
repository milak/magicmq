package conf

import (
	"time"
)
const SIMPLE 	= "SIMPLE"
const VIRTUAL = "VIRTUAL"

/**
 * Stratégies de répartition des topics pour les topics virtuels
 */
const PARAMETER_STRATEGY 	= "Strategy"
const ROUND_ROBIN 			= "ROUND ROBIN"
const ORDERED 				= "ORDERED"

const PARAMETER_DISTRIBUTED = "Distributed"
const DISTRIBUTED_NO 		= "NO"
const DISTRIBUTED_ALL 		= "ALL"
func makeTimestamp() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}
type Topic struct {
	TimeStamp 	int64 	// last update time will be used to resolve synchonisation conflict between instances 
	Name 		string
	Type 		string
	TopicList 	[]string `json:"Topics,omitempty"`
	Parameters 	[]Parameter `json:"Parameters,omitempty"`
}
func NewTopic(aName string) *Topic {
	return &Topic{TimeStamp : makeTimestamp(), Name : aName, Type : SIMPLE}
}
func NewVirtualTopic(aName string, aStrategy string, aTopicList []string) *Topic {
	result := Topic{Name : aName, Type : VIRTUAL, TopicList : aTopicList}
	result.Parameters = make([]Parameter,1)
	result.Parameters[0].Name = PARAMETER_STRATEGY
	result.Parameters[0].Value = ORDERED
	return &result
}
func (this *Topic) IsDistributed() bool {
	for _,parameter := range this.Parameters {
		if parameter.Name == PARAMETER_DISTRIBUTED {
			return parameter.Value != DISTRIBUTED_NO
		}
	}
	return false
}
func (this *Topic) GetParameterByName(aParameterName string) *string {
	for _,p := range this.Parameters {
		if p.Name == aParameterName {
			return &p.Value
		}
	}
	return nil
}