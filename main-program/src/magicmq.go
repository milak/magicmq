package main 

import (
    "flag"
    "fmt"
    "mmq/env"
    "mmq/conf"
    "mmq/service"
    "mmq/item"
    "mmq/dist"
    "time"
)

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var configurationFileName *string = flag.String("f", "configuration.json", "The configuration file name")
func main() {
	context := env.NewContext()
	
    flag.Parse() // Scan the arguments list 
	fmt.Println("Starting MagicMQ on "+context.Host+"...")
    if *versionFlag {
        fmt.Println("Version:"/**, configuration.Version*/)
    }
    
    
    context.Configuration 	= conf.InitConfiguration(*configurationFileName)
	pool 	:= dist.NewInstancePool(context)  
    store 	:= item.NewStore(context)
    itemProcessorService := service.NewItemProcessorService(context,pool)
    itemProcessorService.Start()
    httpService := service.NewHttpService(context,store)
    httpService.Start()
    syncService := service.NewSyncService(context,pool)
    syncService.Start()
    listener := dist.NewListener(context,pool)
    listener.Start()
    autoCleanService := service.NewAutoCleanService(context,store)
    autoCleanService.Start()
    fmt.Println("MagicMQ started")
    for context.Running {
    	time.Sleep(100 * time.Millisecond)
    }
}