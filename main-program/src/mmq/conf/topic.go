package conf

import (

)
const SIMPLE = "SIMPLE"
const VIRTUAL = "VIRTUAL"

/**
 * Stratégies de répartition des topics pour les topics virtuels
 */
const ROUND_ROBIN 	= "ROUND ROBIN"
const ORDERED 		= "ORDERED"
const STRATEGY 		= "strategy"
const PARAMETER_DISTRIBUTED = "Distributed"
const DISTRIBUTED_NO 		= "NO"
const DISTRIBUTED_ALL 		= "ALL" 
type Topic struct {
	Name string
	Type string
	TopicList []string `json:"Topics,omitempty"`
	Parameters 	[]Parameter `json:"Parameters,omitempty"`
}
func NewTopic(aName string) *Topic {
	return &Topic{Name : aName, Type : SIMPLE}
}
func NewVirtualTopic(aName string, aStrategy string, aTopicList []string) *Topic {
	result := Topic{Name : aName, Type : VIRTUAL, TopicList : aTopicList}
	result.Parameters = make([]Parameter,1)
	result.Parameters[0].Name = STRATEGY
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