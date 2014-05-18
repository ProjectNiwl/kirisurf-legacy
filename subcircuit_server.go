// subcircuit_server.go
package main

import (
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"net"

	"github.com/KirisurfProject/kilog"
)

func sc_server_handler(wire io.ReadWriteCloser) (err error) {
	defer func() {
		if err != nil {
			kilog.Debug("sc_server_handler returning err=%s", err.Error())
		}
	}()
	defer wire.Close()
	wire, err = kiss.Obfs3fHandshake(wire, true)
	if err != nil {
		//kilog.Debug(err.Error())
		return nil
	}
	wire, err = kiss.TransportHandshake(MasterKey, wire,
		func([]byte) bool { return true })
	if err != nil {
		kilog.Debug("failed the transport handshake")
		return err
	}
	thing := make([]byte, 1)
	_, err = io.ReadFull(wire, thing)
	if err != nil {
		kilog.Debug("failed the readfull")
		return err
	}
	if thing[0] == 0 {
		// Terminate
		if !MasterConfig.General.IsExit {
			return nil
		}
		e2e_server_handler(wire)
	} else {
		xaxa := make([]byte, thing[0])
		_, err := io.ReadFull(wire, xaxa)
		if err != nil {
			return err
		}
		key := string(xaxa)
		qqq := dirclient.PKeyLookup(key)
		if qqq == nil {
			kilog.Debug("Cannot find %s in %v", xaxa, dirclient.KDirectory)
			return nil
		}
		kilog.Debug("Continuing to %s", qqq.Address)
		rem, err := net.Dial("tcp", qqq.Address)
		if err != nil {
			return err
		}
		remm, err = kiss.Obfs3fHandshake(rem, false)
		if err != nil {
			return err
		}
		go func() {
			io.Copy(wire, remm)
			wire.Close()
		}()
		io.Copy(remm, wire)
		remm.Close()
	}
	return io.EOF
}

type SCServer struct {
	listener net.Listener
	killer   chan bool
}

func NewSCServer(addr string) SCServer {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	killer := make(chan bool)
	go func() {
		for {
			select {
			case <-killer:
				listener.Close()
				return
			default:
				// establish connection
				client, err := listener.Accept()
				//log.Debug("Of acceptings client: %s", client.RemoteAddr())
				if err != nil {
					CRITICAL(err.Error())
					client.Close()
					continue
				}
				go func() {
					err := sc_server_handler(client)
					if err != nil {
						//log.Error(err.Error())
					}
				}()
			}
		}
	}()
	return SCServer{listener, killer}
}

func (thing SCServer) Kill() {
	thing.killer <- true
}
