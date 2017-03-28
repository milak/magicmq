package web

import (
	//"log"
	"mmq/types"
    "net/http"
    "encoding/json"
    "mmq/item"
    "strings"
)
var configuration *types.Configuration
var store *item.ItemStore
func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Sorry "+string(http.StatusNotFound)+" error : not found"))
}
func methodNotSupported(w http.ResponseWriter, aMethod string) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Sorry "+string(http.StatusMethodNotAllowed)+" error : method '"+aMethod+"' not supported"))
}
func infoListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(struct{Version string;IP string}{Version : configuration.Version, IP : "127.0.0.1"})
}
func topicListListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(configuration.Topics)
}
func itemListener(w http.ResponseWriter, req *http.Request){
	if req.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
	} else {
		methodNotSupported(w,req.Method)
	}
}
func topicListener(w http.ResponseWriter, req *http.Request) {
	topicName := req.URL.Path;
	topicName = topicName[len("/topic/"):];
	if req.Method == http.MethodGet {
		if strings.HasSuffix(topicName,"/pop") {
			topicName = topicName[0:len("/pop")]
			item := store.Pop(topicName)
			if item == nil {
				notFound(w)
			} else {
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				encoder.Encode(item)
			}
		} else {
			found := false
			for i := range configuration.Topics {
				topic := configuration.Topics[i]
				if (topic.Name == topicName){
					w.WriteHeader(http.StatusOK)
					encoder := json.NewEncoder(w)
					encoder.Encode(topic)
					found = true
					break;
				}
			}
			if !found {
				notFound(w)
			}
		}
	} else {
		methodNotSupported(w,req.Method)
	}
}
func instanceListListener(w http.ResponseWriter, req *http.Request){
	encoder := json.NewEncoder(w)
	encoder.Encode(configuration.Instances)
	w.WriteHeader(http.StatusOK)
}
func shutdownListener(w http.ResponseWriter, req *http.Request){
	w.WriteHeader(http.StatusOK)
	//http.DefaultServeMux.Shutdown() 
}
func StartHttpListener(aConfiguration *types.Configuration, aStore *item.ItemStore){
	store = aStore
	configuration = aConfiguration
	var port *string = nil
	for s := range configuration.Services {
		service := configuration.Services[s]
		if !service.Active continue
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
			http.HandleFunc("/instance", instanceListListener)
		    http.HandleFunc("/topic", topicListListener)
		    http.HandleFunc("/topic/", topicListener)
		    http.HandleFunc("/item", itemListener)
		    http.HandleFunc("/info", infoListener)
		    http.HandleFunc("/shutdown", shutdownListener)
		}
	}
	if port != nil {
		http.ListenAndServe(":"+(*port), nil)
	}
}