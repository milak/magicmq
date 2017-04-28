package item

import (
	"io"
	"math"
	"time"
	"github.com/google/uuid"
)


/*type Item struct {
	io.Reader
	ID 				string
	Topics	 		[]string
	Reset()
	Properties() 	[]Property
	AddProperty(aName,aValue string) *Property
}*/
type Property struct {
	Name string
	Value string
}
type Item struct {
	ID 				string
	Topics			[]string
	creationDate	time.Time
	value 			[]byte
	ptr 			int
	Properties 		[]Property
	shared			bool
}
func NewItem (aValue []byte, aTopics []string) *Item{
	return &Item{ID : uuid.New().String(), creationDate : time.Now(), value : aValue, Topics : aTopics, ptr : 0}
}
func (this *Item) AddProperty(aName,aValue string) *Property {
	result := Property{Name : aName, Value : aValue}
	this.Properties = append(this.Properties,result)
	return &result
}
func (this *Item) Read(dest []byte) (n int, err error) {
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
func (this *Item) GetAge() time.Duration {
	now := time.Now()
	return now.Sub(this.creationDate)
}
func (this *Item) Size() int {
	return len(this.value)
}
func (this *Item) Reset() {
	this.ptr = 0
}
func (this *Item) SetShared(isShared bool){
	this.shared = isShared
}
func (this *Item) IsShared() bool {
	return this.shared
}