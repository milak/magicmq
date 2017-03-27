package main 

import (
    "flag"
    "fmt"
/*    "io"
    "log"*/
    "mmq/types"
    "mmq/web"
    "mmq/item"
)

var configuration *types.Configuration = types.InitConfiguration()

var store *item.ItemStore = item.NewStore()

func appendTopic(aTopic *types.Topic){
	configuration.Topics = append(configuration.Topics,aTopic)
}
// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
func pop(topic string){
	fmt.Println("Topic : "+topic+" ( nb total = ",store.Count(topic),")")
    item := store.Pop(topic)
    if item == nil {
    	//fmt.Println("plus d'items à lire")
    } else {
	    buffer := make([]byte,1000)
	    count,err := item.Read(buffer)
	    for err == nil {
		    fmt.Println("Item trouvé : ID =",item.ID(),"taille =",count,"value =",string(buffer[0:count]))
		    count,err = item.Read(buffer)
	    }
	    
    }
}
func main() {
    flag.Parse() // Scan the arguments list 
	fmt.Println("Starting MagicMQ...")
    if *versionFlag {
        fmt.Println("Version:", configuration.APP_VERSION)
    }
    //fmt.Println("nb total ",store.Count("test"))
    store.Push(item.NewMemoryItem([]byte("Salut, ceci est un texte de quelques octets"),[]string{"test"}))
    store.Push(item.NewMemoryItem([]byte("Un autre texte"),[]string{"test"}))
    store.Push(item.NewMemoryItem([]byte("Un texte de quelques mots"),[]string{"toto"}))
    appendTopic(types.NewTopic("test"))
    appendTopic(types.NewTopic("tutu"))
    appendTopic(types.NewTopic("toto"))
    appendTopic(types.NewVirtualTopic("v-toto-tutu",[]string{"tutu","toto"}))
    configuration.Instances.Add(types.NewInstance("192.168.0.5","1789"))
    configuration.Instances.Add(types.NewInstance("192.168.0.4","1789"))
    web.Listen(configuration,store)
}