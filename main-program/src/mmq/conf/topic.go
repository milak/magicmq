package conf

import (
	"time"
)
const SIMPLE 	= "SIMPLE"
const VIRTUAL 	= "VIRTUAL"

const PARAMETER_STORE 		= "Store"
const PARAMETER_STORE_RAM	= "RAM"
const PARAMETER_STORE_FS	= "FS"
const PARAMETER_STORE_SWAP	= "SWAP"

const PARAMETER_DISTRIBUTION_STRATEGY = "DistributionStrategy"
const PARAMETER_DISTRIBUTION_STRATEGY_AT_LEAST_ONCE = "AT_LEAST_ONCE"
const PARAMETER_DISTRIBUTION_STRATEGY_EXACTLY_ONCE = "EXACTLY_ONCE"
const PARAMETER_DISTRIBUTION_STRATEGY_AT_MOST_ONCE = "AT_MOST_ONCE"
/**
 * Stratégies de répartition des topics pour les topics virtuels
 */
const PARAMETER_STRATEGY 	= "Strategy"
const ROUND_ROBIN 			= "ROUND_ROBIN"
const ORDERED 				= "ORDERED"

const PARAMETER_DISTRIBUTED = "Distributed"
const DISTRIBUTED_NO 		= "NO"
const DISTRIBUTED_ALL 		= "ALL"

const PARAMETER_DISTRIBUTED_GROUPS = "DistributedGroups"

const PARAMETER_MAX_ITEM_COUNT = "MaxItemCount"

const PARAMETER_TIME_TO_LIVE = "TimeToLive"

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
			return (parameter.Value != DISTRIBUTED_NO && parameter.Value != "")
		}
	}
	return false
}
func (this *Topic) GetParameterByName(aParameterName string) string {
	for _,p := range this.Parameters {
		if p.Name == aParameterName {
			return p.Value
		}
	}
	return ""
}
func (this *Topic) HasParameter(aParameterName string) bool {
	for _,p := range this.Parameters {
		if p.Name == aParameterName {
			return true
		}
	}
	return false
}
func (this *Topic) GetTimeToLive() (*time.Duration, error) {
	if !this.HasParameter(PARAMETER_TIME_TO_LIVE) {
		return nil,nil
	}
	timeToLive := this.GetParameterByName(PARAMETER_TIME_TO_LIVE)
	duration, err := time.ParseDuration(timeToLive)
	if err != nil {
		return nil,err
	}
	return &duration, nil
}
