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

var configuration *types.Configuration

var store *item.ItemStore = item.NewStore()

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var configurationFileName *string = flag.String("f", "configuration.json", "The configuration file name")
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
        fmt.Println("Version:", configuration.Version)
    }
    configuration = types.InitConfiguration(*configurationFileName)
    //fmt.Println("nb total ",store.Count("test"))
    /*store.Push(item.NewMemoryItem([]byte("Salut, ceci est un texte de quelques octets"),[]string{"test"}))
    store.Push(item.NewMemoryItem([]byte("Un autre texte"),[]string{"test"}))
    store.Push(item.NewMemoryItem([]byte("Un texte de quelques mots"),[]string{"toto"}))
    configuration.AddTopic(types.NewTopic("test"))
    configuration.AddTopic(types.NewTopic("tutu"))
    configuration.AddTopic(types.NewTopic("toto"))
    configuration.AddTopic(types.NewVirtualTopic("v-toto-tutu",types.ORDERED,[]string{"tutu","toto"}))
    configuration.AddInstance(types.NewInstance("192.168.0.5","1789"))
    configuration.AddInstance(types.NewInstance("192.168.0.4","1789"))*/
    web.StartSyncListener(configuration,store)
    web.StartHttpListener(configuration,store)
}