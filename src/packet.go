package main

const (
	KeepAlive = iota
	Login
	Handshake
	Chat
	UpdateTime
	PlayerInventory
	SpawnPosition
	_
	_
	_
	Flying
	PlayerPosition
	PlayerLook
	PlayerMoveLook
	BlockDig
	Place
	BlockItemSwitch
	AddToInventory
	ArmAnimation
	_
	NamedEntitySpawn
	PickupSpawn
	CollectItem
	AddObjectOrVehicle
	MobSpawn
	_
	_
	_
	_
	_
	_
	_
	PreChunk
	MapChunk
	MultiBlockChange
	BlockChange
)


type PlayerPositionData struct {
	X, Y, Stance, Z float64
}
type PlayerLookData struct {
	Rotation, Pitch float32
}

type BlockPositionData struct {
	X, Y, Z int32
}

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

type SSpawnPosition struct {
	Position BlockPositionData
}

type CFlying struct {
	IsFlying bool
}

type CPlayerPosition struct {
	Position PlayerPositionData
	Unk      int8
}

type CPlayerLook struct {
	Look PlayerLookData
	Unk  int8
}

type CPlayerMoveLook struct {
	Position PlayerPositionData
	Look     PlayerLookData
	Unk      int8
}

type SPlayerMoveLook struct {
	Position PlayerPositionData
	Look     PlayerLookData
	Unk      int8
}

type SPreChunk struct {
	X, Z int32
	Mode bool
}
type SMapChunk struct {
	X                   int32
	Y                   int16
	Z                   int32
	SizeX, SizeY, SizeZ int8
	Chunk               []byte
}
