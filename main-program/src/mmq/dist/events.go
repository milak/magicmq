package dist

import (
	"mmq/conf"
)
/**
 * Events raised by dist package 
 */
type TopicReceived struct {
	Topic 		*conf.Topic
	From 		*conf.Instance
}
type InstanceReceived struct {
	Instance 	*conf.Instance
	From 		*conf.Instance
}
type InstanceDisconnected struct {
	Instance 	*conf.Instance
}
type ItemReceived struct {
	Item 		*SharedItem
	From		*conf.Instance
}