// Support for reading and writing Named Binary Tags

// FIXME: get rid of tagreader/tagwriter in favor of plain functions

package nbt

import "minecraft/error"

import "compress/gzip"
import "fmt"
import "io"
import "math"
import "os"

type TagType int8

type NamedTag struct {
	Type TagType
	Name string
}

const (
	End TagType = iota
	Byte
	Short
	Int
	Long
	Float
	Double
	ByteArray
	String
	List
	Compound
)

// Load and Save are very common operations that deserve helper functions.

// It would be slightly more correct to take an io.Reader, but this is a convenience
// function anyway.
func Load(file string) (name string, payload map[string]interface{}, err os.Error) {
	gz, err := os.Open(file, os.O_RDONLY, 0000)
	if err != nil {
		err = error.NewError("could not open file", err)
		return
	}
	defer gz.Close()
	nbtf, err := gzip.NewReader(gz)
	if err != nil {
		err = error.NewError("could not gunzip file", err)
		return
	}
	defer nbtf.Close()
	name, payload, err = ReadTagCompound(nbtf)
	if err != nil {
		err = error.NewError("could not read compound tag", err)
		return
	}
	return
}
// It would be slightly more correct to take an io.Writer, but this is a convenience
// function anyway.
func Save(file string, name string, payload map[string]interface{}) (err os.Error) {
	panic("writeme")
}

// Named tag readers.

func ReadNamedTag(reader io.Reader) (t NamedTag, err os.Error) {
	var tag int8
	if tag, err = ReadInt8(reader); err != nil {
		err = error.NewError("could not read tag type", err)
		return
	}
	t.Type = TagType(tag)
	if t.Type == End {
		// end tags have no name; not even a bytelen of 0 for name
		return
	}
	if t.Name, err = ReadString(reader); err != nil {
		return
	}
	return
}

func WriteNamedTag(writer io.Writer, t NamedTag) (err os.Error) {
	panic("writeme")
}


func ReadTagCompound(reader io.Reader) (name string, payload map[string]interface{}, err os.Error) {
	var tag NamedTag
	if tag, err = ReadNamedTag(reader); err != nil {
		err = error.NewError("could not read named tag", err)
		return
	}
	name = tag.Name
	if tag.Type != Compound {
		err = (os.ErrorString)(fmt.Sprint("nbt.ReadTagCompound: expected compound type, got ", tag.Type))
		return
	}
	if payload, err = ReadCompound(reader); err != nil {
		error.NewError("could not read compound tag", err)
		return
	}
	return
}

func readPayload(reader io.Reader, ttype TagType) (payload interface{}, err os.Error) {
	switch ttype {
	case End:
		err = (os.ErrorString)("nbt.readPayload: tag type End has no payload")
	case Byte:
		payload, err = ReadInt8(reader)
		if err != nil {
			err = error.NewError("could not read payload byte", err)
		}
	case Short:
		payload, err = ReadInt16(reader)
		if err != nil {
			err = error.NewError("could not read payload short", err)
		}
	case Int:
		payload, err = ReadInt32(reader)
		if err != nil {
			err = error.NewError("could not read payload int", err)
		}
	case Long:
		payload, err = ReadInt64(reader)
		if err != nil {
			err = error.NewError("could not read payload long", err)
		}
	case Float:
		payload, err = ReadFloat32(reader)
		if err != nil {
			err = error.NewError("could not read payload float", err)
		}
	case Double:
		payload, err = ReadFloat64(reader)
		if err != nil {
			err = error.NewError("could not read payload double", err)
		}
	case ByteArray:
		payload, err = ReadByteArray(reader)
		if err != nil {
			err = error.NewError("could not read payload byte array", err)
		}
	case String:
		payload, err = ReadString(reader)
		if err != nil {
			err = error.NewError("could not read payload string", err)
		}
	case List:
		payload, err = ReadList(reader)
		if err != nil {
			err = error.NewError("could not read payload list", err)
		}
	case Compound:
		payload, err = ReadCompound(reader)
		if err != nil {
			err = error.NewError("could not read payload compound", err)
		}
	default:
		err = (os.ErrorString)(fmt.Sprint("nbt.readPayload: unknown payload type ", ttype))
	}
	return
}

// Payload readers.
// Useful on their own because the Minecraft wire protocol uses the same payload format that nbt files do.

func ReadBool(reader io.Reader) (b bool, err os.Error) {
	var boolSByte int8
	if boolSByte, err = ReadInt8(reader); err != nil {
		err = error.NewError("could not read boolean value", err)
		return
	}
	b = boolSByte != 0 // FIXME? ensure always 0 or 1?
	return
}

func WriteBool(writer io.Writer, b bool) (err os.Error) {
	var boolSByte int8
	if b {
		boolSByte = 1
	} else {
		boolSByte = 0
	}
	if err = WriteInt8(writer, boolSByte); err != nil {
		err = error.NewError("could not write boolean value", err)
		return
	}
	return
}

func ReadByteArray(reader io.Reader) (b []byte, err os.Error) {
	// Normally, we'd use the reader's buffer to read into.  However, it doesn't
	// buy us much here, because it's essentially just scratch space and we need
	// to return something that won't change after we return it.
	var length int32
	if length, err = ReadInt32(reader); err != nil {
		err = error.NewError("could not read byte array's length", err)
		return
	}
	if length < 0 {
		err = error.NewError("byte array's length cannot be < 0", nil)
	}
	b = make([]byte, length)
	if _, err = io.ReadFull(reader, b); err != nil {
		err = error.NewError("could not read byte array", err)
	}
	return
}


func WriteByteArray(writer io.Writer, b []byte) (err os.Error) {
	length := len(b)
	if length > math.MaxInt32 {
		return (os.ErrorString)("nbt.WriteByteArray: byte array was too long")
	}
	if err = WriteInt32(writer, int32(length)); err != nil {
		return
	}
	if _, err = writer.Write(b); err != nil {
		return
	}
	return
}


func ReadCompound(reader io.Reader) (c map[string]interface{}, err os.Error) {
	c = make(map[string]interface{})
	var tag NamedTag
	for {
		if tag, err = ReadNamedTag(reader); err != nil {
			err = error.NewError("could not read named tag", err)
			return
		}
		if tag.Type == End {
			return
		}
		if c[tag.Name], err = readPayload(reader, tag.Type); err != nil {
			err = error.NewError("could not read payload", err)
			return
		}
	}
	panic("shouldn't get here")
}

func ReadFloat32(reader io.Reader) (f float32, err os.Error) {
	var i32 int32
	if i32, err = ReadInt32(reader); err != nil {
		return
	}
	f = math.Float32frombits(uint32(i32))
	return
}

func WriteFloat32(writer io.Writer, f float32) (err os.Error) {
	ui32 := math.Float32bits(f)
	if err = WriteInt32(writer, int32(ui32)); err != nil {
		return
	}
	return
}

func ReadFloat64(reader io.Reader) (f float64, err os.Error) {
	var i64 int64
	if i64, err = ReadInt64(reader); err != nil {
		return
	}
	f = math.Float64frombits(uint64(i64))
	return
}

func WriteFloat64(writer io.Writer, f float64) (err os.Error) {
	ui64 := math.Float64bits(f)
	if err = WriteInt64(writer, int64(ui64)); err != nil {
		return
	}
	return
}

func ReadInt8(reader io.Reader) (i int8, err os.Error) {
	var bytes [1]byte
	if _, err = io.ReadFull(reader, bytes[0:]); err != nil {
		return
	}
	i = int8(bytes[0])
	return
}

func WriteInt8(writer io.Writer, i int8) (err os.Error) {
	var bytes [1]byte

	bytes[0] = byte(i)

	if _, err = writer.Write(bytes[0:]); err != nil {
		return
	}

	return
}

func ReadInt16(reader io.Reader) (i int16, err os.Error) {
	var bytes [2]byte
	if _, err = io.ReadFull(reader, bytes[0:]); err != nil {
		return
	}
	i = int16(uint16(bytes[1]) | uint16(bytes[0])<<8)
	return
}

func WriteInt16(writer io.Writer, i int16) (err os.Error) {
	var bytes [2]byte
	ui := uint16(i)

	bytes[0] = byte(ui >> 8)
	bytes[1] = byte(ui)

	if _, err = writer.Write(bytes[0:]); err != nil {
		return
	}

	return
}

func ReadInt32(reader io.Reader) (i int32, err os.Error) {
	var bytes [4]byte
	if _, err = io.ReadFull(reader, bytes[0:]); err != nil {
		return
	}
	i = int32(uint32(bytes[3]) | uint32(bytes[2])<<8 | uint32(bytes[1])<<16 | uint32(bytes[0])<<24)
	return
}

func WriteInt32(writer io.Writer, i int32) (err os.Error) {
	var bytes [4]byte
	ui := uint32(i)

	bytes[0] = byte(ui >> 24)
	bytes[1] = byte(ui >> 16)
	bytes[2] = byte(ui >> 8)
	bytes[3] = byte(ui)

	if _, err = writer.Write(bytes[0:]); err != nil {
		return
	}

	return
}

func ReadInt64(reader io.Reader) (i int64, err os.Error) {
	var bytes [8]byte
	if _, err = io.ReadFull(reader, bytes[0:]); err != nil {
		return
	}
	i = int64(uint64(bytes[7]) | uint64(bytes[6])<<8 | uint64(bytes[5])<<16 | uint64(bytes[4])<<24 | uint64(bytes[3])<<32 | uint64(bytes[2])<<40 | uint64(bytes[1])<<48 | uint64(bytes[0])<<56)
	return
}

func WriteInt64(writer io.Writer, i int64) (err os.Error) {
	var bytes [8]byte
	ui := uint64(i)

	bytes[0] = byte(ui >> 56)
	bytes[1] = byte(ui >> 48)
	bytes[2] = byte(ui >> 40)
	bytes[3] = byte(ui >> 32)
	bytes[4] = byte(ui >> 24)
	bytes[5] = byte(ui >> 16)
	bytes[6] = byte(ui >> 8)
	bytes[7] = byte(ui)

	if _, err = writer.Write(bytes[0:]); err != nil {
		return
	}

	return
}

func ReadList(reader io.Reader) (l []interface{}, err os.Error) {
	var ttypei8 int8
	var llen int32

	if ttypei8, err = ReadInt8(reader); err != nil {
		err = error.NewError("could not read list type", err)
		return
	}
	if llen, err = ReadInt32(reader); err != nil {
		err = error.NewError("could not read list length", err)
		return
	}
	if llen < 0 {
		err = error.NewError("list length cannot be < 0", nil)
		return
	}
	ttype := TagType(ttypei8)
	// FIXME: we need to make ReadListInt, ReadListCompound, etc...
	l = make([]interface{}, int(llen))
	for i := int32(0); i < llen; i++ {
		var payload interface{}
		if payload, err = readPayload(reader, ttype); err != nil {
			err = error.NewError(fmt.Sprint("could not read list payload at index", i), nil)
			return
		}
		l[i] = payload
	}
	return
}

func ReadString(reader io.Reader) (s string, err os.Error) {
	var strlen int16

	if strlen, err = ReadInt16(reader); err != nil {
		return
	}
	if strlen < 0 {
		err = error.NewError("string length cannot be < 0", nil)
		return
	}
	var strchars = make([]byte, strlen)
	if _, err = io.ReadFull(reader, strchars); err != nil {
		return
	}
	s = string(strchars)
	return
}

func WriteString(writer io.Writer, s string) (err os.Error) {
	var strlenui int16
	var strlen int

	strlen = len(s)
	if strlen > math.MaxInt16 {
		return (os.ErrorString)("nbt.WriteString: string was too long")
	}
	strlenui = (int16)(strlen)
	if err = WriteInt16(writer, strlenui); err != nil {
		return
	}
	strchars := ([]byte)(s)
	if _, err = writer.Write(strchars); err != nil {
		return
	}
	s = string(strchars)
	return
}
