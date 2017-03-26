package main 

import (
    "flag"
    "fmt"
/*    "io"
    "log"*/
    "mmq/types"
    "mmq/web"
)

var configuration types.Configuration = types.InitConfiguration()


func appendTopic(aTopic *types.Topic){
	configuration.Topics = append(configuration.Topics,aTopic)
}
// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")

func main() {
    flag.Parse() // Scan the arguments list 
	fmt.Println("Starting MagicMQ...")
    if *versionFlag {
        fmt.Println("Version:", configuration.APP_VERSION)
    }
    appendTopic(types.NewTopic("test"))
    appendTopic(types.NewTopic("tutu"))
    appendTopic(types.NewTopic("toto"))
    appendTopic(types.NewVirtualTopic("v-toto-tutu",[]string{"tutu","toto"}))
    configuration.Instances.Add(types.NewInstance("www.instance.fr","9898"))
    web.Listen(configuration)
}