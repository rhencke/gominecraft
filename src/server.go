package main

import "net"
import "os"
// import "fmt"

type addClientMsg struct {
	client   *client
	clientId chan int32
}

type server struct {
	listener     net.Listener
	done         chan bool
	clientIdPool chan int32
	addClient    chan *addClientMsg
	clients      []*client
	id           string
	name         string
	motd         string
}

func newServer(address *net.TCPAddr) (s *server, err os.Error) {
	s = new(server)
	l, err := net.ListenTCP("tcp", address)

	if l == nil {
		return nil, err
	}

	s.listener = l
	s.done = make(chan bool)
	s.clientIdPool = make(chan int32, MAX_CLIENTS)
	s.addClient = make(chan *addClientMsg)
	s.clients = make([]*client, MAX_CLIENTS)
	s.id = "abcdef"

	go s.acceptClients()
	go s.clientManager()

	return s, err
}

func (s *server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic("could not accept client." + err.String())
		}
		c := newClient(s, conn)
		go c.accept()
		go s.clientTalk(c)
	}
}

func (s *server) clientManager() {
	go func() {
		for i := (int32)(0); i < MAX_CLIENTS; i++ {
			s.clientIdPool <- i
		}
	}()
	for {
		select {
		case newguy := <-s.addClient:
			cid := <-s.clientIdPool
			s.clients[cid] = newguy.client
			newguy.clientId <- cid
		}
	}
}


func (s *server) clientTalk(client *client) {
	<-client.chandshake
	client.shandshake <- &SHandshake{ServerID: s.id}

	cl := <-client.clogin
	client.Username = cl.Username
	connected := addClientMsg{client, make(chan int32)}
	s.addClient <- &connected
	id := <-connected.clientId
	client.slogin <- &SLogin{PlayerID: id, ServerName: s.name, MOTD: s.motd}
	cl = nil
	for {
		select {
		case <-client.ckeepalive:
			// fmt.Print("kept alive!\n")
		case <-client.cflying:
			// fmt.Print("they flew!\n")
		case <-client.cplayermovelook:
			// fmt.Print("they moved!\n")
		}
	}
}
