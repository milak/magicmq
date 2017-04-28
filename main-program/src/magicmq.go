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
func createServices(context *env.Context, store *item.ItemStore, pool *dist.InstancePool) []service.Service {
	result := []service.Service{}
	result = append(result,service.NewItemProcessorService(context,pool))
	result = append(result,service.NewHttpService(context,store))
	result = append(result,service.NewSyncService(context,pool))
	result = append(result,dist.NewListener(context,pool))
	result = append(result,service.NewAutoCleanService(context,store))
	return result
}
func startServices(services []service.Service){
	for _,service := range services {
		service.Start()
	}
}
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
    
    services := createServices(context,store,pool)
    startServices(services)
    fmt.Println("MagicMQ started")
    for context.Running {
    	time.Sleep(1000 * time.Millisecond)
    }
    fmt.Println("MagicMQ stopped")
}