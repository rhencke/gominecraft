package main

import "minecraft/world"

import "net"
import "os"

type server struct {
	listener  net.Listener
	done      chan bool
	clientMgr *clientMgr
	worldMgr  *worldMgr
	id        string
	name      string
	motd      string
}

func newServer(address *net.TCPAddr, world *world.World) (s *server, err os.Error) {
	s = new(server)
	l, err := net.ListenTCP("tcp", address)

	if l == nil {
		return nil, err
	}

	s.listener = l
	s.done = make(chan bool)
	s.clientMgr = makeClientMgr()
	s.worldMgr = makeWorldMgr(world)
	s.id = "bcfd241a420f886e"

	go s.acceptConnections()
	go s.clientMgr.run(s)
	go s.worldMgr.run(s)
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
		go s.talk(c)
	}
}

func (s *server) talk(conn *conn) {
	<-conn.chandshake
	conn.shandshake <- &SHandshake{ServerID: s.id}
	cl := <-conn.clogin
	idchan := make(chan int32)
	s.clientMgr.addClient <- &addClientReq{&client{cl.Username, conn}, idchan}
	conn.id = <-idchan
	conn.slogin <- &SLogin{PlayerID: conn.id, ServerName: s.name, MOTD: s.motd}
	cl, idchan = nil, nil

	spawnpos := make(chan *BlockPositionData)
	s.worldMgr.getSpawnPos <- &spawnPosReq{conn.id, spawnpos}
	conn.sspawnposition <- &SSpawnPosition{*<-spawnpos}
	for {
		select {
		case <-conn.ckeepalive:
			// fmt.Print("kept alive!\n")
		case <-conn.cflying:
			// fmt.Print("they flew!\n")
		case pml := <-conn.cplayermovelook:
			s.worldMgr.setPlayerPos <- &playerPos{ID: conn.id, Position: &pml.Position}
			s.worldMgr.setPlayerLook <- &playerLook{ID: conn.id, Look: &pml.Look}
		}
	}
}
