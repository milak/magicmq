package service

import (
	"bytes"
	"encoding/json"
	"github.com/milak/event"
	"io"
	"mmq/conf"
	"mmq/env"
	"mmq/item"
	"net/http"
	"os"
	"strings"
	"time"
)

type HttpService struct {
	context *env.Context
	port    string
	store   *item.ItemStore
}

func NewHttpService(aContext *env.Context, aStore *item.ItemStore) *HttpService {
	return &HttpService{context: aContext, store: aStore}
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
		Groups  []string
	}{Version: this.context.Configuration.Version, Host: this.context.Host, Port: this.port, Name: this.context.Host + ":" + this.port, Groups: this.context.Configuration.Groups})
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
	this.context.Logger.Println("DEBUG entering item REST API",req);
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if req.Method == http.MethodPut || req.Method == http.MethodPost {
		//req.ParseForm()
		multipart := true
		erro := req.ParseMultipartForm(http.DefaultMaxHeaderBytes)
		if erro != nil {
			multipart = false
		}
		// Processing topic
		if len(req.Form["topic"]) == 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			if len(req.Form["topic"]) == 0 {
				w.Write([]byte("Missing topic argument"))
			}
			return
		}
		topics := []string{}
		for _, topicName := range req.Form["topic"] {
			topics = append(topics, topicName)
		}
		// Processing value
		var content io.Reader
		if multipart {
			file,_,err := req.FormFile("value")
			content = file
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			defer file.Close()
			
		} else {
			if len(req.Form["value"]) == 0 {
				w.WriteHeader(http.StatusNotAcceptable)
				w.Write([]byte("Missing value argument"))
				return
			}
			content = bytes.NewBuffer([]byte(req.Form["value"][0]))
		}
		item := item.NewItem(topics)
		for i, key := range req.Form["property-name"] {
			value := req.Form["property-value"][i]
			item.AddProperty(key, value)
		}
		this.context.Logger.Println("Adding item");
		err := this.store.Push(item,content)
		if err != nil {
			this.context.Logger.Println("Failed to add",err);
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		} else {
			this.context.Logger.Println("Item added");
			w.WriteHeader(http.StatusCreated)
		}
	} else {
		this.context.Logger.Println("WARNING Method not supported ",req.Method);
		this.methodNotSupported(w, req.Method)
	}
}
type DisplayableItem struct {
	ID 			string
	Age 		time.Duration
	Properties 	[]item.Property
}
func (this *HttpService) topicListener(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	topicName := req.URL.Path
	topicName = topicName[len("/topic/"):]
	if req.Method == http.MethodGet {
		if strings.HasSuffix(topicName, "/pop") {
			topicName = topicName[0 : len(topicName)-len("/pop")]
			item, reader, err := this.store.Pop(topicName)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			} else if item == nil {
				this.notFound(w)
			} else {
				buffer := make([]byte, 1000)
				w.Header().Add("id", item.ID)
				properties := "["
				for i, p := range item.Properties {
					if i != 0 {
						properties += ","
					}
					properties += "{\"name\" : \"" + p.Name + "\", \"value\" : \"" + p.Value + "\"}"
				}
				properties += "]"
				w.Header().Add("properties", properties)
				item.Reset()
				if reader != nil {
					count, err := reader.Read(buffer)
					// TODO support BIG FILE
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write(buffer[0:count])
					}
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			}
		} else if strings.HasSuffix(topicName, "/list") {
			topicName = topicName[0 : len(topicName)-len("/list")]
			items,_ := this.store.List(topicName)
			var displayableItems []DisplayableItem
			for _,i := range items {
				displayableItems = append(displayableItems,DisplayableItem{ID : i.ID, Age : i.GetAge(), Properties : i.Properties})
			}
			this.context.Logger.Println("DEBUG items list of "+topicName,displayableItems)
			w.WriteHeader(http.StatusOK)
			callback := req.Form["callback"]
			this.context.Logger.Println("DEBUG callback",callback)
			if callback != nil {
				w.Write([]byte(callback[0] + "("))
			}
			encoder := json.NewEncoder(w)
			encoder.Encode(displayableItems)
			if callback != nil {
				w.Write([]byte(")"))
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
func (this *HttpService) logListener(w http.ResponseWriter, req *http.Request) {
	file, err := os.Open("mmq.log")
	if err != nil {
		this.notFound(w)
		this.context.Logger.Println("Unable to open log file",err)
	} else {
		w.WriteHeader(http.StatusOK)
		data := make([]byte, 100)
		count, err := file.Read(data)
		if err != nil {
			this.context.Logger.Println("Unable to open log file",err)
		} else {
			for count > 0 {
				w.Write(data[:count])
				count, err = file.Read(data)
			}
		}
		file.Close()
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
			this.context.Logger.Println("Starting ADMIN service with root '"+ (*root)+"'")
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
			http.HandleFunc("/log", this.logListener)
			http.HandleFunc("/shutdown", this.shutdownListener)
		}
	}
	if port != nil {
		this.port = *port
		go http.ListenAndServe(":"+this.port, nil)
	}
}
func (this *HttpService) Stop() {
}