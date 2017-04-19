package dist

import (
	"mmq/env"
	"log"
	"net"
)
type Listener struct {
	listener 	net.Listener					// a reference to the listener when doListen has been called
	context		*env.Context
	pool		*InstancePool
	logger		*log.Logger
	port		string
	running		bool
	protocol	*protocol
}
func NewListener(aContext *env.Context, aPool *InstancePool) *Listener {
	return &Listener{context : aContext, pool : aPool, running : true, protocol : NewProtocol(aContext), logger : aContext.Logger}
}
func (this *Listener) Start(){
	for s := range this.context.Configuration.Services {
		service := this.context.Configuration.Services[s]
		if !service.Active {
			continue
		}
		if service.Name == "SYNC" {
			found := false
			this.logger.Println("starting...")
			for p := range service.Parameters {
				if service.Parameters[p].Name == "port" {
					this.port = service.Parameters[p].Value
					this.running = true
					found = true
					go this.doListen(this.port)
					break
				}
			}
			if !found {
				this.logger.Panic("missing port parameter")
			}
		}
	}
}
/**
 * Listen remote Instances call
 * @param aPort : the listening port
 */
func (this *Listener) doListen (aPort string) {
	this.logger.Println("listening on port",aPort,"...")
	var err error
	this.listener, err = net.Listen("tcp", ":"+aPort)
	if err != nil {
		this.logger.Println("listening failed",err)
	} else {
		for this.running {
			conn, err := this.listener.Accept()
			if err != nil {
				this.logger.Println("Failed to listen",err)
			} else {
				this.logger.Println("caught a call")
				instance,err := this.protocol.handleConnection(conn)
				if err != nil {
					this.pool.newInstanceConnection(instance,&conn)
				}
			}
		}
	}
}
func (this *Listener) Stop() {
	this.running = false
	if this.listener != nil {
		this.listener.Close()
	}
}