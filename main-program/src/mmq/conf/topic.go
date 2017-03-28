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

type Topic struct {
	Name string
	Type string
	TopicList []string `json:"Topics,omitempty"`
	Strategy string `json:"Strategy,omitempty"`
}
func NewTopic(aName string) *Topic {
	return &Topic{Name : aName, Type : SIMPLE}
}
func NewVirtualTopic(aName string, aStrategy string, aTopicList []string) *Topic {
	return &Topic{Name : aName, Type : VIRTUAL, Strategy : aStrategy, TopicList : aTopicList}
}
/**func (slice ByteSlice) Append(data []byte) []byte {
    // Body exactly the same as the Append function defined above.
}*/