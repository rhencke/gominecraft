package main

import "net"
import "fmt"
import "minecraft/nbt"
import "os"

type client struct {
	server          *server
	connection      net.Conn
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
		if pktype, err = nbt.ReadInt8(c.connection); err != nil {
			return
		}

		switch (PacketId)(pktype) {
		case KeepAlive:
			ck := new(CKeepAlive)
			c.ckeepalive <- ck
		case Login:
			cl := new(CLogin)
			if cl.ProtocolVersion, err = nbt.ReadInt32(c.connection); err != nil {
				return
			}
			if cl.Username, err = nbt.ReadString(c.connection); err != nil {
				return
			}
			if cl.Password, err = nbt.ReadString(c.connection); err != nil {
				return
			}
			c.clogin <- cl
		case Handshake:
			ch := new(CHandshake)
			if ch.Username, err = nbt.ReadString(c.connection); err != nil {
				return
			}
			c.chandshake <- ch
		case Flying:
			cf := new(CFlying)
			if cf.IsFlying, err = nbt.ReadBool(c.connection); err != nil {
				return
			}
			c.cflying <- cf
		case PlayerMoveLook:
			cpml := new(CPlayerMoveLook)
			if cpml.X, err = nbt.ReadFloat64(c.connection); err != nil {
				return
			}
			if cpml.Y, err = nbt.ReadFloat64(c.connection); err != nil {
				return
			}
			if cpml.Stance, err = nbt.ReadFloat64(c.connection); err != nil {
				return
			}
			if cpml.Z, err = nbt.ReadFloat64(c.connection); err != nil {
				return
			}
			if cpml.Rotation, err = nbt.ReadFloat32(c.connection); err != nil {
				return
			}
			if cpml.Pitch, err = nbt.ReadFloat32(c.connection); err != nil {
				return
			}
			if cpml.Unk, err = nbt.ReadInt8(c.connection); err != nil {
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
			if err = nbt.WriteInt8(c.connection,int8(Login)); err != nil {
				return
			}
			if err = nbt.WriteInt32(c.connection,sl.PlayerID); err != nil {
				return
			}
			if err = nbt.WriteString(c.connection,sl.ServerName); err != nil {
				return
			}
			if err = nbt.WriteString(c.connection,sl.MOTD); err != nil {
				return
			}
		case sh := <-c.shandshake:
			if err = nbt.WriteInt8(c.connection,int8(Handshake)); err != nil {
				return
			}
			if err = nbt.WriteString(c.connection,sh.ServerID); err != nil {
				return
			}
		}
	}
}
