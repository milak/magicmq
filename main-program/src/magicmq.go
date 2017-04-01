package main 

import (
    "flag"
    "fmt"
    "mmq/env"
    "mmq/conf"
    "mmq/service"
    "mmq/item"
    "time"
)

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var configurationFileName *string = flag.String("f", "configuration.json", "The configuration file name")
/**func pop(topic string){
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
}*/
func main() {
    flag.Parse() // Scan the arguments list 
	fmt.Println("Starting MagicMQ...")
    if *versionFlag {
        fmt.Println("Version:"/**, configuration.Version*/)
    }
    context := *env.NewContext()
    context.Configuration 	= conf.InitConfiguration(*configurationFileName)
    context.Store 			= item.NewStore()
    httpService := service.NewHttpService(&context)
    httpService.Start()
    syncService := service.NewSyncService(&context)
    syncService.Start()
    fmt.Println("MagicMQ started")
    for context.Running {
    	time.Sleep(100 * time.Millisecond)
    }
}