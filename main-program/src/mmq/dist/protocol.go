package dist

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/milak/event"
	"io"
	"log"
	"mmq/conf"
	"mmq/env"
	"net"
	"strconv"
	"time"
)
type protocol struct {
	context	*env.Context
	logger	*log.Logger
	port 	string
}
// A const for DIEZE value
const dieze byte = byte('#')
func NewProtocol(aContext *env.Context) *protocol {
	result := &protocol{context : aContext, logger : aContext.Logger}
	for _,service := range aContext.Configuration.Services {
		if service.Name == conf.SERVICE_SYNC {
			for _,parameter := range service.Parameters {
				if parameter.Name == "port" {
					result.port = parameter.Value
				}
			}
		}
	}
	return result
}
/**
 * Internal method that splits a command received from remote 
 */
func (this *protocol) splitCommand(line []byte) (command string, arguments []byte, remain []byte, needMore int) {
	this.logger.Println("Splitting ",string(line))
	// Synchronize with firts DIEZE
	if line[0] != dieze {
		return "",[]byte{},line,0
	}
	// Obtain the COMMAND
	i := 1
	command = ""
	for line[i] != dieze {
		command += string(line[i])
		i++
	}
	this.logger.Println("Splitting - command found ",command)
	// Obtain the LENGTH
	i++
	slength := ""
	for line[i] != dieze {
		slength += string(line[i])
		i++
	}
	i++
	this.logger.Println("Splitting - slength found ",slength)
	length,_ := strconv.Atoi(slength)
	// Check wether the whole data has been received
	remain = []byte{}
	needMore = 0
	tailleRestantALire := len(line) - i
	// Missing the end the data
	if length > tailleRestantALire { // la longueur de la donnée annoncée > tailleRestantALire
		this.logger.Println("Splitting - not finished arguments found ",string(arguments))
		arguments = line[i:]
		needMore = length - tailleRestantALire // la longueur de la donnée annoncée - la tailleRestantALire + un dieze
		this.logger.Println("Splitting - need more",needMore)
	} else { // All the data has been obtained
		arguments = line[i:i+length]
		this.logger.Println("Splitting - arguments found ",string(arguments))
		// check wether a part of a next command is found in the data received
		if (i+length) == len(line) {
			this.logger.Println("Splitting - nothing remain")
		} else {
			remain = line[i+length:]
			this.logger.Println("Splitting - something remain", string(remain))
		}
	}
	return command, arguments, remain, needMore
}

/**
 * Send a command to remote Instance :
 *   #<command>#<size of arguments>#<arguments>
 * @param command    : a command cannot be nul
 * @param arguments  : a data to send cannot be nul, if empty, send []byte{}
 * @param connection : an open connection
 */
func (this *protocol) sendCommand(command string, arguments []byte, connection net.Conn){
	this.logger.Println("sending",command,string(arguments))
	connection.Write([]byte("#"+command+"#"+strconv.Itoa(len(arguments))+"#"))
	connection.Write(arguments)
}

/**
 * Method used by both side : caller and called
 * all the commands will be received through this link
 */
func (this *protocol) keepConnected(aInstance *conf.Instance, aConnection *net.Conn){
	var byteBuffer *bytes.Buffer
	buffer := make([]byte,2000)
	connection := (*aConnection) 
	defer func() {
		connection.Close()
	}()
	var command string
	var arguments, remain []byte
	var needMore int
	// while the service is running
	for this.context.Running {
		time.Sleep(1 * time.Second)
		// reintroduce remain from previous command
		if len(remain) > 0 {
			this.logger.Println("Reusing remain",remain)
			buffer = remain
		} else {
			this.logger.Println("Listening to remote")
			count,err := connection.Read(buffer)
			if err != nil {
				this.logger.Println("Lost connection with",aInstance.Name(),err)
				event.EventBus.FireEvent(&InstanceDisconnected{Instance : aInstance})
				break
			}
			buffer = buffer[0:count]
		}
		// Parse data received
		command, arguments, remain, needMore = this.splitCommand(buffer)
		if needMore != 0 {
			this.logger.Println("Unfinished",arguments,", need",needMore,"bytes")
			var bufferNeeded bytes.Buffer // Todo in the future use a swap in disk for ITEM command
			remain = this.takeMore(connection,needMore,&bufferNeeded)
			arguments = append(arguments,bufferNeeded.Bytes()...)
			this.logger.Println("Finally got",string(arguments))
		}
		// Process the command
		this.logger.Println("Received command " + command,remain)
		// Received HELLO from called
		if command == "HELLO" { // On est côté appelant, on reçoit la réponse de l'appelé, on lui envoie la configuration
			this.sendConfiguration(&connection)
		} else if command == "INSTANCES" { // Receive instance list
			var newInstances []*conf.Instance
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&newInstances)
			for _,newInstance := range newInstances {
				this.logger.Println("Received instance :",newInstance)
				event.EventBus.FireEvent(&InstanceReceived{Instance : newInstance, From : aInstance})
			}
		} else if command == "TOPICS" { // Receive topic list
			var distributedTopics []*conf.Topic
			byteBuffer = bytes.NewBuffer(arguments)
			decoder := json.NewDecoder(byteBuffer)
			decoder.Decode(&distributedTopics)
			for _,topic := range distributedTopics {
				this.logger.Println("Received topic :",topic)
				event.EventBus.FireEvent(&TopicReceived{Topic : topic, From : aInstance})
			}
		} else if command == "ERROR" {
			this.logger.Println("Received ERROR :",arguments)
		} else {
			this.logger.Println("Not supported command")
			this.sendCommand("ERROR",[]byte("NOT SUPPORTED COMMAND '"+command+"'"),connection)
		}
	}
}
/**
 * Send configuration to other side :
 *   * the known instances
 *   * the distributed topics
 */
func (this *protocol) sendConfiguration(aConnection *net.Conn){
	this.logger.Println("Sending configuration")
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(this.context.Configuration.Instances)
	this.sendCommand("INSTANCES",buffer.Bytes(),*aConnection)
	buffer.Reset()
	var distibutedTopics []*conf.Topic
	for _,topic := range this.context.Configuration.Topics {
		if topic.IsDistributed() {
			distibutedTopics = append(distibutedTopics,topic)
		}
	}
	if len (distibutedTopics) > 0 {
		encoder.Encode(distibutedTopics)
		this.logger.Println("Sending topics ", string(buffer.Bytes()))
		this.sendCommand("TOPICS",buffer.Bytes(),*aConnection)
	}
}
/**
 * Process the connection when called by a remote node.
 */
func (this *protocol) handleConnection (aConn net.Conn) (*conf.Instance, error) {
	this.context.Host,_,_ = net.SplitHostPort(aConn.LocalAddr().String())
	this.logger.Println("Processing call")
	buffer := make([]byte,1000)
	count,err := aConn.Read(buffer)
	if err != nil {
		this.logger.Println("Unable to read HELLO from remote",err)
		return nil,err
	}
	if count < 10 {
		this.logger.Println("Unable to read HELLO from remote ",buffer[0:count])
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return nil,err
	}
	command, arguments, remain, needMore := this.splitCommand(buffer[0:count])
	if needMore > 0 {
		this.logger.Println("Unfinished",arguments,", need",needMore,"bytes")
		var bufferNeeded bytes.Buffer
		remain = this.takeMore(aConn,needMore,&bufferNeeded)
		arguments = append(arguments,bufferNeeded.Bytes()...)
		this.logger.Println("Finally got",string(arguments))
	}
	if command == "" {
		this.logger.Println("Unable to read HELLO from remote")
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return nil,errors.New("Unable to understand")
	}
	this.logger.Println("Received ",command,"-",arguments,"-",remain)
	if command != "HELLO" {
		this.logger.Println("Unable to read HELLO from remote ")
		this.sendCommand("ERROR",[]byte("Unable to understand"),aConn)
		return nil,errors.New("Unable to understand " + command)
	}
	host,port,_ := net.SplitHostPort(string(arguments))
	instance := conf.NewInstance(host,port)
	//this.logger.Println("Adding caller as new instance",instance)
	event.EventBus.FireEvent(&InstanceReceived{Instance : instance, From : nil})
	instance.Connected = true
	for len(remain) > 0 {
		command, arguments, remain, needMore = this.splitCommand(remain)
		this.logger.Println("Received command " + command,arguments,remain,needMore)
	}
	this.sendCommand("HELLO",[]byte(this.context.Host+":"+this.port),aConn) // TODO échanger leur numéros de version
	
	this.sendConfiguration(&aConn)
	go this.keepConnected(instance,&aConn)
	return instance, nil
	// TODO : gerer le fait que les deux peuvent essayer de se connecter en même temps, il y aura alors deux connections entre eux
}
/**
 * Reads from connection until missing data is received (linked to NeedMore detection in splitCommand)
 */
func (this *protocol) takeMore(connection net.Conn, needMore int, writer io.Writer) (remain []byte){
	buffer := make([]byte,2000)
	var total = 0
	for total < needMore {
		count,err := connection.Read(buffer)
		if err != nil {
			this.logger.Println("Failed to read following needed bytes")
			return
		}
		if count + total > needMore {
			part := needMore - total
			writer.Write(buffer[0:part])
			remain = buffer[part:count]
		} else {
			total += count
			writer.Write(buffer[0:count])
			remain = []byte{}
		}
	}
	return remain
}