package service

import (
	"encoding/json"
	"mmq/env"
	"mmq/conf"
	"mmq/item"
	"github.com/milak/event"
	"net/http"
	"strings"
)

type HttpService struct {
	context *env.Context
	port    string
	store   *item.ItemStore
}

func NewHttpService(aContext *env.Context, aStore *item.ItemStore) *HttpService {
	return &HttpService{context: aContext, store : aStore}
}
func (this *HttpService) notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Sorry " + string(http.StatusNotFound) + " error : not found"))
}
func (this *HttpService) methodNotSupported(w http.ResponseWriter, aMethod string) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Sorry " + string(http.StatusMethodNotAllowed) + " error : method '" + aMethod + "' not allowed"))
}
func (this *HttpService) infoListener(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	req.ParseForm()
	callback := req.Form["callback"]
	w.WriteHeader(http.StatusOK)
	if callback != nil {
		w.Write([]byte(callback[0] + "("))
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(struct {
		Version string
		Host    string
		Port    string
		Name    string
	}{Version: this.context.Configuration.Version, Host: this.context.Host, Port: this.port, Name: this.context.Host + ":" + this.port})
	if callback != nil {
		w.Write([]byte(")"))
	}
}
func (this *HttpService) topicListListener(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	req.ParseForm()
	callback := req.Form["callback"]
	w.WriteHeader(http.StatusOK)
	if callback != nil {
		w.Write([]byte(callback[0] + "("))
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(this.context.Configuration.Topics)
	if callback != nil {
		w.Write([]byte(")"))
	}
}
func (this *HttpService) itemListener(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPut || req.Method == http.MethodPost {
		req.ParseForm()
		if len(req.Form["topic"]) == 0 || len(req.Form["value"]) == 0 {
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			topics := []string{}
			for _,topicName := range req.Form["topic"] { 
				topics = append(topics, topicName)
			}
			value := req.Form["value"][0]
			item := item.NewMemoryItem([]byte(value), topics)
			for i,key := range req.Form["property-name"] {
				value := req.Form["property-value"][i]
				item.AddProperty(key,value)
			}
			err := this.store.Push(item)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else {
				w.WriteHeader(http.StatusCreated)
			}
		}
	} else {
		this.methodNotSupported(w, req.Method)
	}
}
func (this *HttpService) topicListener(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	topicName := req.URL.Path
	topicName = topicName[len("/topic/"):]
	if req.Method == http.MethodGet {
		if strings.HasSuffix(topicName, "/pop") {
			topicName = topicName[0 : len(topicName)-len("/pop")]
			item,err := this.store.Pop(topicName)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else if item == nil {
				this.notFound(w)
			} else {
				buffer := make([]byte, 1000)
				w.Header().Add("id", item.ID())
				properties := "["
				for i,p := range item.Properties() {
					if i != 0 {
						properties += ","
					}
					properties += "{\"name\" : \""+p.Name+"\", \"value\" : \""+p.Value+"\"}"
				}
				properties += "]"
				w.Header().Add("properties",properties)
				count, err := item.Read(buffer)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Write(buffer[0:count])
				}
			}
		} else {
			topic := this.context.Configuration.GetTopic(topicName)
			if topic == nil {
				this.notFound(w)
			} else {
				req.ParseForm()
				w.WriteHeader(http.StatusOK)
				callback := req.Form["callback"]
				if callback != nil {
					w.Write([]byte(callback[0] + "("))
				}
				encoder := json.NewEncoder(w)
				encoder.Encode(topic)
				if callback != nil {
					w.Write([]byte(")"))
				}
			}
		}
	} else if req.Method == http.MethodDelete {
		if !this.context.Configuration.RemoveTopic(topicName) {
			this.notFound(w)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		this.methodNotSupported(w, req.Method)
	}
}
func (this *HttpService) instanceListListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(this.context.Configuration.Instances)
}
func (this *HttpService) instanceListener(w http.ResponseWriter, req *http.Request) {
	instanceName := req.URL.Path
	instanceName = instanceName[len("/instance/"):]
	if req.Method == http.MethodDelete {
		removedInstance := this.context.Configuration.RemoveInstance(instanceName)
		if removedInstance == nil {
			this.notFound(w)
		} else {
			event.EventBus.FireEvent(&conf.InstanceRemoved{removedInstance})
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (this *HttpService) shutdownListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	this.context.Running = false
	//http.DefaultServeMux.Shutdown()
}
func (this *HttpService) Start() {
	var port *string = nil
	for s := range this.context.Configuration.Services {
		service := this.context.Configuration.Services[s]
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
			http.HandleFunc("/instance", this.instanceListListener)
			http.HandleFunc("/instance/", this.instanceListener)
			http.HandleFunc("/topic", this.topicListListener)
			http.HandleFunc("/topic/", this.topicListener)
			http.HandleFunc("/item", this.itemListener)
			http.HandleFunc("/info", this.infoListener)
			http.HandleFunc("/shutdown", this.shutdownListener)
		}
	}
	if port != nil {
		this.port = *port
		go http.ListenAndServe(":"+this.port, nil)
	}
}
