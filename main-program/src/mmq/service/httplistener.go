package service

import (
	//"log"
	"mmq/conf"
    "net/http"
    "encoding/json"
    "mmq/item"
    "strings"
)
type HttpService struct {
	configuration 	*conf.Configuration
	store 			*item.ItemStore
}
func NewHttpService (aConfiguration *conf.Configuration, aStore *item.ItemStore) *HttpService {
	return &HttpService{configuration : aConfiguration, store : aStore}
}
func (this *HttpService) notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Sorry "+string(http.StatusNotFound)+" error : not found"))
}
func (this *HttpService) methodNotSupported(w http.ResponseWriter, aMethod string) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Sorry "+string(http.StatusMethodNotAllowed)+" error : method '"+aMethod+"' not allowed"))
}
func (this *HttpService) infoListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(struct{Version string;IP string}{Version : this.configuration.Version, IP : "127.0.0.1"})
}
func (this *HttpService) topicListListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(this.configuration.Topics)
}
func (this *HttpService) itemListener(w http.ResponseWriter, req *http.Request){
	if req.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
	} else {
		this.methodNotSupported(w,req.Method)
	}
}
func (this *HttpService) topicListener(w http.ResponseWriter, req *http.Request) {
	topicName := req.URL.Path;
	topicName = topicName[len("/topic/"):];
	if req.Method == http.MethodGet {
		if strings.HasSuffix(topicName,"/pop") {
			topicName = topicName[0:len("/pop")]
			item := this.store.Pop(topicName)
			if item == nil {
				this.notFound(w)
			} else {
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				encoder.Encode(item)
			}
		} else {
			found := false
			for i := range this.configuration.Topics {
				topic := this.configuration.Topics[i]
				if (topic.Name == topicName){
					w.WriteHeader(http.StatusOK)
					encoder := json.NewEncoder(w)
					encoder.Encode(topic)
					found = true
					break;
				}
			}
			if !found {
				this.notFound(w)
			}
		}
	} else if req.Method == http.MethodDelete {
		if !this.configuration.RemoveTopic(topicName){
			this.notFound(w)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		this.methodNotSupported(w,req.Method)
	}
}
func (this *HttpService) instanceListListener(w http.ResponseWriter, req *http.Request){
	encoder := json.NewEncoder(w)
	encoder.Encode(this.configuration.Instances)
	w.WriteHeader(http.StatusOK)
}
func (this *HttpService) shutdownListener(w http.ResponseWriter, req *http.Request){
	w.WriteHeader(http.StatusOK)
	//http.DefaultServeMux.Shutdown() 
}
func (this *HttpService) Start(){
	var port *string = nil
	for s := range this.configuration.Services {
		service := this.configuration.Services[s]
		if !service.Active {
			continue
		}
		if service.Name == "ADMIN" {
			var root *string = nil
			for p := range service.Parameters {
				if service.Parameters[p].Name == "root" {
					root = &service.Parameters[p].Value
					break
				}
			}
			if root == nil {
				panic("Configuration error : missing root parameter for ADMIN service")
			}
			http.Handle("/", http.FileServer(http.Dir(*root)))
		} else if service.Name == "REST" {
			for p := range service.Parameters {
				if service.Parameters[p].Name == "port" {
					port = &service.Parameters[p].Value
					break
				}
			}
			if port == nil {
				panic("Configuration error : missing port parameter for REST service")
			}
			http.HandleFunc("/instance", 	this.instanceListListener)
		    http.HandleFunc("/topic", 		this.topicListListener)
		    http.HandleFunc("/topic/", 		this.topicListener)
		    http.HandleFunc("/item", 		this.itemListener)
		    http.HandleFunc("/info", 		this.infoListener)
		    http.HandleFunc("/shutdown", 	this.shutdownListener)
		}
	}
	if port != nil {
		http.ListenAndServe(":"+(*port), nil)
	}
}