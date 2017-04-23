package dist

import (
	"mmq/conf"
)
// Event
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
	Item 		*ManagedItem
	From		*conf.Instance
}