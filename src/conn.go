package main

import "minecraft/nbt"

import "bufio"
import "bytes"
import "compress/zlib"
import "fmt"
import "io"
import "net"
import "os"

type conn struct {
	connection      net.Conn
	errs            chan os.Error
	ckeepalive      chan *CKeepAlive
	clogin          chan *CLogin
	slogin          chan *SLogin
	chandshake      chan *CHandshake
	shandshake      chan *SHandshake
	sspawnposition  chan *SSpawnPosition
	cflying         chan *CFlying
	cplayerposition chan *CPlayerPosition
	cplayermovelook chan *CPlayerMoveLook
	splayermovelook chan *SPlayerMoveLook
	sprechunk       chan *SPreChunk
	smapchunk       chan *SMapChunk
	id              int32
}

func newConn(server *server, netconn net.Conn) *conn {
	return &conn{
		netconn,
		make(chan os.Error),
		make(chan *CKeepAlive),
		make(chan *CLogin),
		make(chan *SLogin),
		make(chan *CHandshake),
		make(chan *SHandshake),
		make(chan *SSpawnPosition),
		make(chan *CFlying),
		make(chan *CPlayerPosition),
		make(chan *CPlayerMoveLook),
		make(chan *SPlayerMoveLook),
		make(chan *SPreChunk),
		make(chan *SMapChunk),
		-1}
}

func (c *conn) accept() {
	done := make(chan os.Error)
	go c.read(done)
	go c.write(done)
	err := <-done
	if err != nil {
		fmt.Print(err.String())
	}
	c.connection.Close()
}

func (c *conn) read(done chan<- os.Error) {
	var err os.Error
	defer func() {
		if err != nil {
			c.errs <- err
		}
	}()
	bc := bufio.NewReader(c.connection)
	for {
		var pktype int8
		if pktype, err = nbt.ReadInt8(bc); err != nil {
			return
		}

		switch pktype {
		case KeepAlive:
			ck := new(CKeepAlive)
			c.ckeepalive <- ck
		case Login:
			cl := new(CLogin)
			if cl.ProtocolVersion, err = nbt.ReadInt32(bc); err != nil {
				return
			}
			if cl.Username, err = nbt.ReadString(bc); err != nil {
				return
			}
			if cl.Password, err = nbt.ReadString(bc); err != nil {
				return
			}
			c.clogin <- cl
		case Handshake:
			ch := new(CHandshake)
			if ch.Username, err = nbt.ReadString(bc); err != nil {
				return
			}
			c.chandshake <- ch
		case Flying:
			cf := new(CFlying)
			if cf.IsFlying, err = nbt.ReadBool(bc); err != nil {
				return
			}
			c.cflying <- cf
		case PlayerPosition:
			cpp := new(CPlayerPosition)
			if cpp.Position, err = readPlayerPos(bc); err != nil {
				return
			}
			c.cplayerposition <- cpp
		case PlayerMoveLook:
			cpml := new(CPlayerMoveLook)
			if cpml.Position, err = readPlayerPos(bc); err != nil {
				return
			}
			if cpml.Look.Rotation, err = nbt.ReadFloat32(bc); err != nil {
				return
			}
			if cpml.Look.Pitch, err = nbt.ReadFloat32(bc); err != nil {
				return
			}
			if cpml.Unk, err = nbt.ReadInt8(bc); err != nil {
				return
			}
			c.cplayermovelook <- cpml
		default:
			fmt.Printf("unknown packet type: %d\n", pktype)
		}
	}
	return
}

func (c *conn) write(done chan<- os.Error) {
	var err os.Error
	defer func() {
		if err != nil {
			c.errs <- err
		}
	}()
	bc := bufio.NewWriter(c.connection)

	for {
		select {
		case sl := <-c.slogin:
			if err = nbt.WriteInt8(bc, Login); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, sl.PlayerID); err != nil {
				return
			}
			if err = nbt.WriteString(bc, sl.ServerName); err != nil {
				return
			}
			if err = nbt.WriteString(bc, sl.MOTD); err != nil {
				return
			}
		case sh := <-c.shandshake:
			if err = nbt.WriteInt8(bc, Handshake); err != nil {
				return
			}
			if err = nbt.WriteString(bc, sh.ServerID); err != nil {
				return
			}
		case ssp := <-c.sspawnposition:
			if err = nbt.WriteInt8(bc, SpawnPosition); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, ssp.Position.X); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, ssp.Position.Y); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, ssp.Position.Z); err != nil {
				return
			}
		case spml := <-c.splayermovelook:
			if err = nbt.WriteInt8(bc, PlayerMoveLook); err != nil {
				return
			}
			if err = nbt.WriteFloat64(bc, spml.Position.X); err != nil {
				return
			}
			if err = nbt.WriteFloat64(bc, spml.Position.Y); err != nil {
				return
			}
			if err = nbt.WriteFloat64(bc, spml.Position.Stance); err != nil {
				return
			}
			if err = nbt.WriteFloat64(bc, spml.Position.Z); err != nil {
				return
			}
			if err = nbt.WriteFloat32(bc, spml.Look.Rotation); err != nil {
				return
			}
			if err = nbt.WriteFloat32(bc, spml.Look.Pitch); err != nil {
				return
			}
			if err = nbt.WriteInt8(bc, spml.Unk); err != nil {
				return
			}
		case spc := <-c.sprechunk:
			if err = nbt.WriteInt8(bc, PreChunk); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, spc.X); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, spc.Z); err != nil {
				return
			}
			if err = nbt.WriteBool(bc, spc.Mode); err != nil {
				return
			}
		case sch := <-c.smapchunk:
			if err = nbt.WriteInt8(bc, int8(MapChunk)); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, sch.X); err != nil {
				return
			}
			if err = nbt.WriteInt16(bc, sch.Y); err != nil {
				return
			}
			if err = nbt.WriteInt32(bc, sch.Z); err != nil {
				return
			}
			if err = nbt.WriteInt8(bc, sch.SizeX); err != nil {
				return
			}
			if err = nbt.WriteInt8(bc, sch.SizeY); err != nil {
				return
			}
			if err = nbt.WriteInt8(bc, sch.SizeZ); err != nil {
				return
			}
			// Minecraft expects map chunks to be zlib compressed.
			var b bytes.Buffer
			compressor, err := zlib.NewWriter(&b)
			if err != nil {
				return
			}
			compressor.Write(sch.Chunk)
			compressor.Close()
			if err = nbt.WriteByteArray(bc, b.Bytes()); err != nil {
				return
			}
		}

		// send packet
		if err = bc.Flush(); err != nil {
			return
		}
	}
}

func readPlayerPos(r io.Reader) (p PlayerPositionData, err os.Error) {
	if p.X, err = nbt.ReadFloat64(r); err != nil {
		return
	}
	if p.Y, err = nbt.ReadFloat64(r); err != nil {
		return
	}
	if p.Stance, err = nbt.ReadFloat64(r); err != nil {
		return
	}
	if p.Z, err = nbt.ReadFloat64(r); err != nil {
		return
	}
	return
}
