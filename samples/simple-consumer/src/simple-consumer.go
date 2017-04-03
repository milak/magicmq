package main 

import (
	"flag"
	"fmt"
	//"io/ioutil"
	"net/http"
	"encoding/json"
	"strings"
)
// The flag package provides a server host name
var hostFlag *string = flag.String("h", "", "The server host name")
var portFlag *string = flag.String("p", "", "The server port")
func post(url, 
func requestDelete(url string){
	request,_ := http.NewRequest(http.MethodDelete, url, http.NoBody)
	resp, err := http.DefaultClient.Do(request)
    if err != nil {
    	fmt.Println("Echec lors de l'appel au serveur")
    	return
    }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusOK {
    	fmt.Println("OK")
    } else {
	    fmt.Println("Fail :", resp.Status)
    }
}
func getObject(url string, object interface{}) bool {
	//fmt.Println("Calling ",url)
	resp, err := http.Get(url)
    if err != nil {
    	fmt.Println("Echec lors de l'appel au serveur")
    	return false
    }
    defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println("response ", body)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&object)
	return true
}
func main() {
	flag.Parse() // Scan the arguments list 
	if *hostFlag == "" {
		fmt.Println("Missing host argument")
		flag.Usage()
		return
	}
	if *portFlag == "" {
		fmt.Println("Missing port argument")
		flag.Usage()
		return
	}
	server := "http://"+*hostFlag+":"+*portFlag
	fmt.Println("Starting simple-consumer...")
    information := new (struct {Version string})
    if !getObject(server+"/info",&information) {
    	return
    }
	fmt.Println("Connected to server version n°" + information.Version)
	var command,arg1,arg2,arg3,arg4,arg5,arg6,arg7 string
	var currentDirectory = "/"
	for {
		fmt.Print(currentDirectory+"> ")
		count,_ := fmt.Scanln(&command,&arg1,&arg2,&arg3,&arg4,&arg5,&arg6,&arg7)
		command = strings.Trim(command,"\t ")
		if len(command) == 0 {
			continue
		} else if command == "exit" || command == "bye" || command == "quit" {
			if count != 1 {
				fmt.Println("too much arguments")
				continue
			}
			break
		} else if command == "ls" {
			if count != 1 {
				fmt.Println("Arguments not supported")
				continue
			}
			if currentDirectory == "/" {
				fmt.Println("topic")
				fmt.Println("instance")
				fmt.Println("service")
			} else if currentDirectory == "/topic" {
				topicList := make ([]struct {Name string; Type string},100)
				getObject(server+"/topic",&topicList)
				for i := range topicList {
					topic := topicList[i]
					fmt.Println(topic.Type + "\t" + topic.Name)
				}
			} else if currentDirectory == "/instance" {
				instanceList := make ([]struct {Host string; Port string; Connected bool},100)
				getObject(server+"/instance",&instanceList)
				for _,instance := range instanceList {
					fmt.Println(instance.Host + ":" + instance.Port,instance.Connected)
				}
			} else {
				fmt.Println("Not yet implemented")
			}
		} else if command == "help" {
			if count != 1 {
				fmt.Println("too much arguments")
				continue
			}
			fmt.Println("available commands :")
			fmt.Println(" help : display help")
			fmt.Println(" ls : list folders or items in the current position")
			fmt.Println(" cd <position> : change current position")
			fmt.Println(" mk <object> <options>...: create an object :")
			fmt.Println(" 	item     : mk item -f <file> <topics>... [<parameters>...] | mk item -c <content> <topics> [<parameters>...]")
			fmt.Println(" 	           <topics>     : <topicname>;<topicname>;...")
			fmt.Println(" 	           <parameters> : <parameter>=<parameterValue>;")
			fmt.Println(" 	topic    : mk topic <name> <type> <parameters>...")
			fmt.Println(" 	instance : mk instance <host> <port>")
			fmt.Println(" quit | exit | bye : exit")
		} else if command == "mk" {
			if arg1 == "item" {
				if count < 4 {
					fmt.Println("Not enough argument")
				} else {
					topics := arg4 
					if arg2 == "-f" {
						
					} else if arg2 == "-c" {
						
					} else {
						fmt.Println("Option not supported " + arg2)
					}
				}
			} else if arg1 == "topic" {
				
			} else if arg1 == "instance" {
				
			} else {
				fmt.Println("Can't create " + arg1)
			}
		} else if command == "cd" {
			if count == 1 {
				fmt.Println("missing argument")
				continue
			} else if count != 2 {
				fmt.Println("too much arguments")
				continue
			}
			directory := arg1
			if directory == ".." {
				if directory == "/" {
					continue
				} else {
					pos := strings.LastIndex(currentDirectory,"/")
					if pos == 0 {
						currentDirectory = "/"
					} else {
						currentDirectory = currentDirectory[0:pos-1]
					}
				}
			} else {
				if directory[0:1] != "/" {
					directory = currentDirectory + directory
				}
				folders := []string{"/topic","/instance","/service"}
				found := false
				for d := range folders {
					if directory == folders[d] {
						found = true
						currentDirectory = directory
						break
					}
				}
				if !found {
					fmt.Println("folder "+directory+" not found")
				}
			}
		} else if command == "rm" {
			if count == 1 {
				fmt.Println("missing argument")
				continue
			} else if count != 2 {
				fmt.Println("too much arguments")
				continue
			}
			if currentDirectory == "/topic" {
				var topic = arg1
				requestDelete(server+"/topic/"+topic)
			} else if currentDirectory == "/instance" {
				var instance = arg1
				requestDelete(server+"/instance/"+instance)
			} else if currentDirectory == "/service" {
				
			} 
		} else {
			fmt.Println("unsupported command '" + command+"'" + "'" + command[0:2] +"'")
		}
	}
}