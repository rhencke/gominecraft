package main

type PacketId int8 // yep, it's really signed

const (
	KeepAlive PacketId = iota
	Login
	Handshake
	Chat
	UpdateTime
	_
	_
	_
	_
	_
	Flying
	PlayerPosition
	PlayerLook
	PlayerMoveLook
)

type CKeepAlive struct{}

type SLogin struct {
	PlayerID         int32
	ServerName, MOTD string
}

type CLogin struct {
	ProtocolVersion    int32
	Username, Password string
}

type SHandshake struct {
	ServerID string
}
type CHandshake struct {
	Username string
}

type CFlying struct {
	IsFlying bool
}

type CPlayerMoveLook struct {
	X, Y, Stance, Z float64
	Rotation, Pitch float32
	Unk int8
}
