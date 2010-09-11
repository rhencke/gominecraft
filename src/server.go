package main

import "net"
import "os"
// import "fmt"


type client struct{}

type server struct {
	listener  net.Listener
	done      chan bool
	clientMgr *clientMgr
	id        string
	name      string
	motd      string
}

type clientMgr struct {
	clientIdPool chan int32
	addClient    chan chan int32
	clients      []*client
}

func newServer(address *net.TCPAddr) (s *server, err os.Error) {
	s = new(server)
	l, err := net.ListenTCP("tcp", address)

	if l == nil {
		return nil, err
	}

	s.listener = l
	s.done = make(chan bool)
	s.clientMgr = &clientMgr{
		make(chan int32, MAX_CLIENTS),
		make(chan chan int32),
		make([]*client, MAX_CLIENTS),
	}
	s.id = "abcdef"

	go s.acceptConnections()
	go s.clientMgr.run()

	return s, err
}

func (s *server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic("could not accept client." + err.String())
		}
		c := newConn(s, conn)
		go c.accept()
		go s.clientTalk(c)
	}
}

func (c *clientMgr) run() {
	go func() {
		for i := (int32)(0); i < MAX_CLIENTS; i++ {
			c.clientIdPool <- i
		}
	}()
	for {
		select {
		case newguy := <-c.addClient:
			cid := <-c.clientIdPool
			c.clients[cid] = &client{}
			newguy <- cid
		}
	}
}


func (s *server) clientTalk(conn *conn) {
	<-conn.chandshake
	conn.shandshake <- &SHandshake{ServerID: s.id}

	cl := <-conn.clogin
	conn.Username = cl.Username
	connected := make(chan int32)
	s.clientMgr.addClient <- connected
	id := <-connected
	conn.slogin <- &SLogin{PlayerID: id, ServerName: s.name, MOTD: s.motd}
	cl = nil
	for {
		select {
		case <-conn.ckeepalive:
			// fmt.Print("kept alive!\n")
		case <-conn.cflying:
			// fmt.Print("they flew!\n")
		case <-conn.cplayermovelook:
			// fmt.Print("they moved!\n")
		}
	}
}
