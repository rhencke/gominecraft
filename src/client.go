package main

import "net"
import "fmt"
import "minecraft/nbt"
import "os"

type client struct {
	server          *server
	connection      net.Conn
	nr              *nbt.TagReader
	nw              *nbt.TagWriter
	errs            chan os.Error
	ckeepalive      chan *CKeepAlive
	clogin          chan *CLogin
	slogin          chan *SLogin
	chandshake      chan *CHandshake
	shandshake      chan *SHandshake
	cflying         chan *CFlying
	cplayermovelook chan *CPlayerMoveLook
	Username        string
}

func newClient(server *server, conn net.Conn) *client {
	return &client{server,
		conn,
		nbt.NewTagReader(conn),
		nbt.NewTagWriter(conn),
		make(chan os.Error),
		make(chan *CKeepAlive),
		make(chan *CLogin),
		make(chan *SLogin),
		make(chan *CHandshake),
		make(chan *SHandshake),
		make(chan *CFlying),
		make(chan *CPlayerMoveLook),
		""}
}

func (c *client) accept() {
	done := make(chan os.Error)
	go c.read(done)
	go c.write(done)
	err := <-done
	if err != nil {
		fmt.Print(err.String())
	}
	c.connection.Close()
}

func (c *client) read(done chan<- os.Error) {
	var err os.Error
	defer func() {
		if err != nil {
			c.errs <- err
		}
	}()

	for {
		var pktype int8
		if pktype, err = c.nr.ReadInt8(); err != nil {
			return
		}

		switch (PacketId)(pktype) {
		case KeepAlive:
			ck := new(CKeepAlive)
			c.ckeepalive <- ck
		case Login:
			cl := new(CLogin)
			if cl.ProtocolVersion, err = c.nr.ReadInt32(); err != nil {
				return
			}
			if cl.Username, err = c.nr.ReadString(); err != nil {
				return
			}
			if cl.Password, err = c.nr.ReadString(); err != nil {
				return
			}
			c.clogin <- cl
		case Handshake:
			ch := new(CHandshake)
			if ch.Username, err = c.nr.ReadString(); err != nil {
				return
			}
			c.chandshake <- ch
		case Flying:
			cf := new(CFlying)
			if cf.IsFlying, err = c.nr.ReadBool(); err != nil {
				return
			}
			c.cflying <- cf
		case PlayerMoveLook:
			cpml := new(CPlayerMoveLook)
			if cpml.X, err = c.nr.ReadFloat64(); err != nil {
				return
			}
			if cpml.Y, err = c.nr.ReadFloat64(); err != nil {
				return
			}
			if cpml.Stance, err = c.nr.ReadFloat64(); err != nil {
				return
			}
			if cpml.Z, err = c.nr.ReadFloat64(); err != nil {
				return
			}
			if cpml.Rotation, err = c.nr.ReadFloat32(); err != nil {
				return
			}
			if cpml.Pitch, err = c.nr.ReadFloat32(); err != nil {
				return
			}
			if cpml.Unk, err = c.nr.ReadInt8(); err != nil {
				return
			}
			c.cplayermovelook <- cpml
		default:
			fmt.Printf("unknown packet type: %d\n", pktype)
		}
	}
	return
}

func (c *client) write(done chan<- os.Error) {
	var err os.Error
	defer func() {
		if err != nil {
			c.errs <- err
		}
	}()
	for {
		select {
		case sl := <-c.slogin:
			if err = c.nw.WriteInt8(int8(Login)); err != nil {
				return
			}
			if err = c.nw.WriteInt32(sl.PlayerID); err != nil {
				return
			}
			if err = c.nw.WriteString(sl.ServerName); err != nil {
				return
			}
			if err = c.nw.WriteString(sl.MOTD); err != nil {
				return
			}
		case sh := <-c.shandshake:
			if err = c.nw.WriteInt8(int8(Handshake)); err != nil {
				return
			}
			if err = c.nw.WriteString(sh.ServerID); err != nil {
				return
			}
		}
	}
}
