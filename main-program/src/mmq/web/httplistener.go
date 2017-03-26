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
	io.WriteString(w, "{\n")
	io.WriteString(w, "\ttopics  : [\n")
	for i := range configuration.Topics {
		io.WriteString(w, "\t\ttopic : {\n\t\t\tName : \"" + configuration.Topics[i].Name + "\"\n\t\t}\n")
	}
	io.WriteString(w, "\t]\n")
	io.WriteString(w, "}\n")
}

func topicListener(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\n")
	topicName := req.URL.Path;
	topicName = topicName[len("/topic/"):];
	for i := range configuration.Topics {
		topic := configuration.Topics[i]
		if (topic.Name == topicName){
			io.WriteString(w, "\t\ttopic : {\n");
			io.WriteString(w, "\t\t\tName : \"" + topic.Name + "\",\n");
			if (topic.Type == types.SIMPLE){
				io.WriteString(w, "\t\t\tType : \"SIMPLE\",\n");
			} else if (topic.Type == types.VIRTUAL){
				io.WriteString(w, "\t\t\tType : \"VIRTUAL\",\n");
				io.WriteString(w, "\t\t\tList : [\n");
				for l := range topic.TopicList {
					 io.WriteString(w, "\t\t\t\"" + topic.TopicList[l] + "\"");
				}
				io.WriteString(w, "\t\t\t]\n");
			}
			io.WriteString(w, "\t\t}\n")
			break;
		}
	}
	io.WriteString(w, "}\n")
}
func instanceListListener(w http.ResponseWriter, req *http.Request){
	
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