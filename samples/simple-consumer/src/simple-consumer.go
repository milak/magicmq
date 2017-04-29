package main

import (
	"bytes"
	"flag"
	"fmt"
	//"io/ioutil"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)
type NameValue struct {
	Name string
	Value string
}
// The flag package provides a server host name
var hostFlag *string = flag.String("h", "", "The server host name")
var portFlag *string = flag.String("p", "", "The server port")

func post(url string, data url.Values) {
	resp, err := http.PostForm(url, data)
	if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
		return
	}
	if resp.StatusCode == http.StatusCreated {
		fmt.Println("OK")
	} else {

		fmt.Println("Fail", http.StatusText(resp.StatusCode))
	}
}
func requestDelete(url string) {
	request, _ := http.NewRequest(http.MethodDelete, url, http.NoBody)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("OK")
	} else {
		fmt.Println("Fail :", resp.Status)
	}
}
func get(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
		return false
	}
	defer resp.Body.Close()
	buffer := make([]byte, 1000)
	count, err := resp.Body.Read(buffer)
	for k, v := range resp.Header {
		fmt.Println(k, v)
	}
	/*if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
		return false
	}*/
	for count > 0 {
		fmt.Print(string(buffer[0:count]))
		count, err = resp.Body.Read(buffer)
		/*if err != nil {
			fmt.Println("Echec lors de l'appel au serveur", err)
			return false
		}*/
	}
	fmt.Println()
	return true
}
func getObject(url string, object interface{}) bool {
	//fmt.Println("Calling ",url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
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
	server := "http://" + *hostFlag + ":" + *portFlag
	fmt.Println("Starting simple-consumer...")
	information := new(struct {
		Host    string
		Port    string
		Version string
		Groups  []string
	})
	if !getObject(server+"/info", &information) {
		return
	}
	fmt.Println("Connected to server version n°" + information.Version)
	var command, arg1, arg2, arg3, arg4, arg5, arg6, arg7 string
	var currentDirectory = "/"
	for {
		fmt.Print(currentDirectory + "> ")
		count, _ := fmt.Scanln(&command, &arg1, &arg2, &arg3, &arg4, &arg5, &arg6, &arg7)
		command = strings.Trim(command, "\t ")
		if len(command) == 0 {
			continue
		} else if command == "info" {
			fmt.Println("Server information")
			fmt.Println(" - host    :", information.Host)
			fmt.Println(" - port    :", information.Port)
			fmt.Println(" - version :", information.Version)
			fmt.Println(" - groups  :", information.Groups)
		} else if command == "exit" || command == "bye" || command == "quit" {
			if count != 1 {
				fmt.Println("too much arguments")
				continue
			}
			break
		} else if command == "log" {
			get(server + "/log")
		} else if command == "ls" {
			if count != 1 {
				fmt.Println("Arguments not supported")
				continue
			}
			commandLs(server, currentDirectory)
		} else if command == "help" {
			if count != 1 {
				fmt.Println("too much arguments")
				continue
			}
			commandHelp()
		} else if command == "mk" {
			if arg1 == "item" {
				if count < 4 {
					fmt.Println("Not enough argument")
				} else {
					topics := arg4
					var content []byte
					if arg2 == "-f" {
						content = []byte(arg3) // TODO read the file
					} else if arg2 == "-c" {
						content = []byte(arg3)
					} else {
						fmt.Println("Option not supported " + arg2)
					}
					values := make(url.Values)
					values["topic"] = []string{topics}
					values["value"] = []string{string(content)}
					post(server+"/item", values)
				}
			} else if arg1 == "topic" {

			} else if arg1 == "instance" {

			} else {
				fmt.Println("Can't create " + arg1)
			}
		} else if command == "pop" {
			commandPop(server, currentDirectory, count, arg1)
		} else if command == "cd" {
			currentDirectory = commandCD(server, currentDirectory, count, arg1)
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
				requestDelete(server + "/topic/" + topic)
			} else if currentDirectory == "/instance" {
				var instance = arg1
				requestDelete(server + "/instance/" + instance)
			} else if currentDirectory == "/service" {

			}
		} else {
			fmt.Println("unsupported command '" + command + "'")
		}
	}
}
func fill(text string, what string, max int) string {
	result := text
	total := len(text)
	for total < max {
		result += what
		total++
	}
	return result
}
func commandPop(server string, currentDirectory string, count int, arg1 string) {
	var topic string
	if count == 1 { // must see if we are in topic
		if strings.HasPrefix(currentDirectory, "/topic/") {
			topic = currentDirectory[len("/topic/"):]
		} else {
			fmt.Println("MIssing topic argument or must be in topic")
			return
		}
	} else if count == 2 { // must see if we are in topic
		topic = arg1
	} else {
		fmt.Println("Too much arguments")
		return
	}
	resp, err := http.Get(server + "/topic/" + topic + "/pop")
	if err != nil {
		fmt.Println("Echec lors de l'appel au serveur", err)
		return
	}
	largestName := len("Name")
	largestValue := len("Value")
	propList := make([]NameValue,0)
	ID := resp.Header["Id"][0]
	if len(ID) > largestValue {
		largestValue = len(ID)
	}
	propList = append(propList,NameValue{Name : "ID", Value : ID})
	// Other properties
	propertiesJSON := resp.Header["Properties"][0]
	bytesBuffer := bytes.NewBuffer([]byte(propertiesJSON))
	decoder := json.NewDecoder(bytesBuffer)
	properties := make([]NameValue, 0)
	decoder.Decode(&properties)
	for _,property := range properties {
		propList = append(propList,NameValue{Name : property.Name, Value : property.Value})
		if len(property.Name) > largestName {
			largestName = len(property.Name)
		}
		if len(property.Value) > largestValue {
			largestValue = len(property.Value)
		}
	}
	// Display
	fmt.Println("--"+fill("","-",largestName)+"---"+fill("","-",largestValue)+"--")
	fmt.Println("| "+fill("Name"," ",largestName)+" | "+fill("Value"," ",largestValue)+" |")
	fmt.Println("--"+fill("","-",largestName)+"---"+fill("","-",largestValue)+"--")
	for _,nameValue := range propList {
		fmt.Println("| "+fill(nameValue.Name," ",largestName)+" | "+fill(nameValue.Value," ",largestValue)+" |")
	}
	fmt.Println("--"+fill("","-",largestName)+"---"+fill("","-",largestValue)+"--")
	defer resp.Body.Close()
	fmt.Println("Content :")
	buffer := make([]byte, 1000)
	countBytes, err := resp.Body.Read(buffer)
	for countBytes > 0 {
		fmt.Print(string(buffer[0:countBytes]))
		countBytes, err = resp.Body.Read(buffer)
	}
	fmt.Println()
}
func commandCD(server string, currentDirectory string, count int, arg1 string) string {
	if count == 1 {
		fmt.Println("missing argument")
		return currentDirectory
	} else if count != 2 {
		fmt.Println("too much arguments")
		return currentDirectory
	}
	directory := arg1
	if directory == ".." {
		if directory == "/" {
			return currentDirectory
		} else {
			pos := strings.LastIndex(currentDirectory, "/")
			if pos == 0 {
				currentDirectory = "/"
			} else {
				currentDirectory = currentDirectory[0:pos]
			}
		}
	} else {
		if directory[0:1] != "/" {
			if currentDirectory == "/" {
				directory = currentDirectory + directory
			} else {
				directory = currentDirectory + "/" + directory
			}
		}
		folders := []string{"/topic", "/instance", "/service"}
		// add topics
		topicList := make([]struct {
			Name string
			Type string
		}, 0)
		getObject(server+"/topic", &topicList)
		for _, topic := range topicList {
			folders = append(folders, "/topic/"+topic.Name)
		}
		found := false
		for d := range folders {
			if directory == folders[d] {
				found = true
				currentDirectory = directory
				break
			}
		}
		if !found {
			fmt.Println("folder " + directory + " not found")
		}
	}
	return currentDirectory
}
func commandLs(server string, currentDirectory string) {
	if currentDirectory == "/" {
		fmt.Println("topic")
		fmt.Println("instance")
		fmt.Println("service")
	} else if currentDirectory == "/topic" {
		topicList := make([]struct {
			Name string
			Type string
		}, 0)
		getObject(server+"/topic", &topicList)
		largestTopicName := len("Topic")
		largestTopicType := len("Type")
		for _, topic := range topicList {
			if len(topic.Name) > largestTopicName {
				largestTopicName = len(topic.Name)
			}
			if len(topic.Type) > largestTopicType {
				largestTopicType = len(topic.Type)
			}
		}
		fmt.Println("--" + fill("", "-", largestTopicName) + "---" + fill("", "-", largestTopicType) + "--")
		fmt.Println("| " + fill("Topic", " ", largestTopicName) + " | " + fill("Type", " ", largestTopicType) + " |")
		fmt.Println("--" + fill("", "-", largestTopicName) + "---" + fill("", "-", largestTopicType) + "--")
		for _, topic := range topicList {
			fmt.Println("|", fill(topic.Name, " ", largestTopicName), "|", fill(topic.Type, " ", largestTopicType), "|")
		}
		fmt.Println("--" + fill("", "-", largestTopicName) + "---" + fill("", "-", largestTopicType) + "--")
	} else if currentDirectory == "/instance" {
		instanceList := make([]struct {
			Host      string
			Port      string
			Connected bool
			Groups    []string
			GroupList string
		}, 0)
		getObject(server+"/instance", &instanceList)
		largestHost := len("Host")
		largestGroupList := len("Groups")
		for idx, instance := range instanceList {
			if len(instance.Host) > largestHost {
				largestHost = len(instance.Host)
			}
			gl := ""
			for i, group := range instance.Groups {
				if i > 0 {
					gl += ", "
				}
				gl += group
			}
			instanceList[idx].GroupList = gl
			if len(gl) > largestGroupList {
				largestGroupList = len(gl)
			}
			fmt.Println(gl)
		}
		fmt.Println("--" + fill("", "-", largestHost) + "----------------------" + fill("", "-", largestGroupList) + "--")
		fmt.Println("| " + fill("Host", " ", largestHost) + " | Port | Connected | " + fill("Groups", " ", largestGroupList) + " |")
		fmt.Println("--" + fill("", "-", largestHost) + "----------------------" + fill("", "-", largestGroupList) + "--")
		for _, instance := range instanceList {
			var connected string
			if instance.Connected {
				connected = "true     "
			} else {
				connected = "false    "
			}
			fmt.Println("|", fill(instance.Host, " ", largestHost), "|", instance.Port, "|", connected, "|", fill(instance.GroupList, " ", largestGroupList), "|")
		}
		fmt.Println("--" + fill("", "-", largestHost) + "----------------------" + fill("", "-", largestGroupList) + "--")
	} else if strings.HasPrefix(currentDirectory, "/topic/") {
		topic := currentDirectory[len("/topic/"):]
		fmt.Println("Items in '" + topic + "'")
		itemList := make([]struct {
			ID         string
			Age        int
			AgeString  string
			Properties []NameValue
			PropertiesAsLine string
		}, 0)
		getObject(server+"/topic/"+topic+"/list", &itemList)
		largestID := len("ID")
		largestAge := len("Age")
		largestProperties := len("Properties")
		for idx, item := range itemList {
			if len(item.ID) > largestID {
				largestID = len(item.ID)
			}
			itemList[idx].AgeString = strconv.Itoa(item.Age) + "ms"
			if len(itemList[idx].AgeString) > largestAge {
				largestAge = len(itemList[idx].AgeString)
			}
			var properties string
			for y, property := range item.Properties {
				if y > 0 {
					properties += ", "
				}
				properties += property.Name + " : " + property.Value
			}
			itemList[idx].PropertiesAsLine = properties
			if len(properties) > largestProperties {
				largestProperties = len(properties)
			}
		}
		fmt.Println("--" + fill("", "-", largestID) + "---" + fill("", "-", largestAge) + "---" + fill("", "-", largestProperties) + "--")
		fmt.Println("| " + fill("ID", " ", largestID) + " | " + fill("Age", " ", largestAge) + " | " + fill("Properties", " ", largestProperties) + " |")
		fmt.Println("--" + fill("", "-", largestID) + "---" + fill("", "-", largestAge) + "---" + fill("", "-", largestProperties) + "--")
		for _, item := range itemList {
			fmt.Println("|", fill(item.ID, " ", largestID), "|", fill(item.AgeString, " ", largestAge), "|", fill(item.PropertiesAsLine, " ", largestProperties), "|")
		}
		fmt.Println("--" + fill("", "-", largestID) + "---" + fill("", "-", largestAge) + "---" + fill("", "-", largestProperties) + "--")
	} else {
		fmt.Println("Not yet implemented")
	}
}
func commandHelp() {
	fmt.Println("available commands :")
	fmt.Println(" help : display help")
	fmt.Println(" info : display informations about current server")
	fmt.Println(" log  : display logs of the current serveur")
	fmt.Println(" ls   : list folders or items in the current position")
	fmt.Println(" cd <position> : change current position")
	fmt.Println(" mk <object> <options>...: create an object :")
	fmt.Println(" 	item     : mk item -f <file> <topics>... [<parameters>...] | mk item -c <content> <topics> [<parameters>...]")
	fmt.Println(" 	           <topics>     : <topicname>;<topicname>;...")
	fmt.Println(" 	           <parameters> : <parameter>=<parameterValue>;")
	fmt.Println(" 	topic    : mk topic <name> <type> <parameters>...")
	fmt.Println(" 	instance : mk instance <host> <port>")
	fmt.Println(" pop [topic]  : get an item from topic, if current directory in topic, argument topic is optional")
	fmt.Println(" quit | exit | bye : exit")
}
