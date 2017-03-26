package web

import (
	"io"
	"log"
	"mmq/types"
    "net/http"
    "encoding/json"
)
var configuration types.Configuration 
func infoListener(w http.ResponseWriter, req *http.Request) {
	
	io.WriteString(w, "{\n")
	io.WriteString(w, "\tinstance : {\n")
	io.WriteString(w, "\t\tversion : \""+configuration.APP_VERSION+"\"\n")
	io.WriteString(w, "\t\tIP : \"127.0.0.1\"\n")
	io.WriteString(w, "\t\tport : \"8080\"\n")
	io.WriteString(w, "\t}\n")
	io.WriteString(w, "}\n")
}
func topicListListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	encoder := json.NewEncoder(w)
	encoder.Encode(configuration.Topics)
}

func topicListener(w http.ResponseWriter, req *http.Request) {
	topicName := req.URL.Path;
	topicName = topicName[len("/topic/"):];
	found := false
	for i := range configuration.Topics {
		topic := configuration.Topics[i]
		if (topic.Name == topicName){
			encoder := json.NewEncoder(w)
			encoder.Encode(topic)
			found = true
			break;
		}
	}
	if !found {
		w.WriteHeader(404)
	}
}
func instanceListListener(w http.ResponseWriter, req *http.Request){
	encoder := json.NewEncoder(w)
	encoder.Encode(configuration.Instances)
}
func Listen(aConfiguration types.Configuration){
	configuration = aConfiguration
	http.HandleFunc("/instance", instanceListListener)
    http.HandleFunc("/topic", topicListListener)
    http.HandleFunc("/topic/", topicListener)
    http.HandleFunc("/info", infoListener)
    http.Handle("/", http.FileServer(http.Dir(configuration.WebDirectory)))
    log.Fatal(http.ListenAndServe(":"+configuration.WebAdminPort, nil))
}