package main 

import (
    "flag"
    "fmt"
    "net/http"
    "io"
    "log"
    "mmq/types"
)

const APP_VERSION = "0.1"
var topics []*types.Topic
func InfoListener(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\n")
	io.WriteString(w, "\tinstance : {\n")
	io.WriteString(w, "\t\tversion : \""+APP_VERSION+"\"\n")
	io.WriteString(w, "\t\tIP : \"127.0.0.1\"\n")
	io.WriteString(w, "\t\tport : \"8080\"\n")
	io.WriteString(w, "\t}\n")
	io.WriteString(w, "}\n")
}
func TopicListListener(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	io.WriteString(w, "{\n")
	io.WriteString(w, "\ttopics  : [\n")
	for i := range topics {
		io.WriteString(w, "\t\ttopic : {\n\t\t\tName : \"" + topics[i].Name + "\"\n\t\t}\n")
	}
	io.WriteString(w, "\t]\n")
	io.WriteString(w, "}\n")
}
func TopicListener(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\n")
	topicName := req.URL.Path;
	topicName = topicName[len("/topic/"):];
	for i := range topics {
		topic := topics[i]
		if (topic.Name == topicName){
			io.WriteString(w, "\t\ttopic : {\n");
			io.WriteString(w, "\t\t\tName : \"" + topic.Name + "\",\n");
			if (topic.Type == types.SIMPLE){
				io.WriteString(w, "\t\t\tType : \"SIMPLE\",\n");
			} else if (topic.Type == types.VIRTUAL){
				io.WriteString(w, "\t\t\tType : \"VIRTUAL\",\n");
				io.WriteString(w, "\t\t\tList : [\n");
				for l := range topic.TopicList {
					 io.WriteString(w, "\t\t\t\"" + topic.TopicList[l].Name + "\"");
				}
				io.WriteString(w, "\t\t\t]\n");
			}
			io.WriteString(w, "\t\t}\n")
			break;
		}
	}
	io.WriteString(w, "}\n")
}
func appendTopic(aTopic *types.Topic){
	topics = append(topics,aTopic)
}
// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")

func main() {
    flag.Parse() // Scan the arguments list 
	fmt.Println("Starting MagicMQ...")
    if *versionFlag {
        fmt.Println("Version:", APP_VERSION)
    }
    appendTopic(types.NewTopic("test"))
    appendTopic(types.NewTopic("tutu"))
    appendTopic(types.NewTopic("toto"))
    appendTopic(types.NewVirtualTopic("v-toto-tutu",[]string("tutu","toto")))
    http.HandleFunc("/topic", TopicListListener)
    http.HandleFunc("/topic/", TopicListener)
    http.HandleFunc("/info", InfoListener)
    log.Fatal(http.ListenAndServe(":8080", nil))
}