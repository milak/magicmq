package main 

import (
	"flag"
	"fmt"
	//"io/ioutil"
	"net/http"
	"encoding/json"
)
// The flag package provides a server host name
var hostFlag *string = flag.String("h", "", "The server host name")
var portFlag *string = flag.String("p", "", "The server port")

func main() {
	flag.Parse() // Scan the arguments list 
	if *hostFlag == "" {
		flag.Usage()
		return
	}
	fmt.Println("Starting simple-consumer...")
	
    resp, err := http.Get("http://"+*hostFlag+":"+*portFlag+"/info")
    if err != nil {
    	fmt.Println("Echec lors de l'appel au serveur")
    }
    defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println("response ", body)
	decoder := json.NewDecoder(resp.Body)
	information := new (struct {Version string})
	decoder.Decode(&information)
	fmt.Println(information.Version)
}