package types

import (

)
const SIMPLE = "SIMPLE"
const VIRTUAL = "VIRTUAL"

type Topic struct {
	Name string
	Type string
	TopicList []string
}
func NewTopic(aName string) *Topic {
	return &Topic{Name : aName, Type : SIMPLE}
}
func NewVirtualTopic(aName string, aTopicList []string) *Topic {
	return &Topic{Name : aName, Type : VIRTUAL, TopicList : aTopicList}
}
/**func (slice ByteSlice) Append(data []byte) []byte {
    // Body exactly the same as the Append function defined above.
}*/