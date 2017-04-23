package conf

import (

)
const SERVICE_ADMIN = "ADMIN"
const SERVICE_REST 	= "REST"
const SERVICE_SYNC 	= "SYNC"
type Service struct {
	Name 		string
	Comment 	string
	Active 		bool
	Parameters 	[]Parameter `json:"Parameters,omitempty"`
}
func (this *Service) GetParameterByName(aName string) *Parameter{
	for _,p := range this.Parameters {
		if p.Name == aName {
			return &p
		}
	}
	return nil
}