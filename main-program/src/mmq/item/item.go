package item

import (
	"io"
	"math"
	"github.com/google/uuid"
)
type Item interface {
	io.Reader
	ID() 			string
	Topics() 		[]string
	Reset()
	Properties() 	[]Property
}
type Property struct {
	Name string
	Value string
}
type memoryItem struct {
	id 			string
	topics		[]string
	value 		[]byte
	ptr 		int
	properties 	[]Property
}
func (this *memoryItem) ID() string {
	return this.id
}
func (this *memoryItem) Properties() []Property {
	return this.properties
}
func (this *memoryItem) Topics() []string {
	return this.topics
}
func (this *memoryItem) Read(dest []byte) (n int, err error) {
	if this.ptr >= len(this.value) {
		return 0,io.EOF
	} else {
		reste := len(this.value) - this.ptr
		count := int(math.Min(float64(reste),float64(len(dest))))
		copy(dest,this.value[this.ptr:this.ptr+count])
		this.ptr = this.ptr + count
		return count,nil
	}
}
func (this *memoryItem) Reset() {
	this.ptr = 0
}
func NewMemoryItem (aValue []byte, aTopics []string) Item{
	return &memoryItem{id : uuid.New().String(), value : aValue, topics : aTopics, ptr : 0}
}