// Support for reading and writing Named Binary Tags

package nbt

import "io"
import "os"
import "math"
import "minecraft/error"
import "fmt"

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

type TagReader struct {
	r   io.Reader
	buf [512]byte
}

type TagWriter struct {
	w   io.Writer
	buf [8]byte
}

// func Read(r io.Reader) (data map[string]interface{}, err os.Error) {
// 	return readTagCompound(r)
// }

func NewTagReader(reader io.Reader) *TagReader {
	return &TagReader{r: reader}
}

func NewTagWriter(writer io.Writer) *TagWriter {
	return &TagWriter{w: writer}
}

// Named tag readers.

func (reader *TagReader) ReadNamedTag() (t NamedTag, err os.Error) {
	var tag int8
	if tag, err = reader.ReadInt8(); err != nil {
		err = error.NewError("could not read string length", err)
		return
	}
	t.Type = TagType(tag)
	if t.Type == End {
		// end tags have no name; not even a bytelen of 0 for name
		return
	}
	if t.Name, err = reader.ReadString(); err != nil {
		return
	}
	return
}

func (writer *TagWriter) WriteNamedTag(t NamedTag) (err os.Error) {
	panic("writeme")
}


func (reader *TagReader) ReadTagCompound() (name string, payload map[string]interface{}, err os.Error) {
	var tag NamedTag
	if tag, err = reader.ReadNamedTag(); err != nil {
		err = error.NewError("could not read named tag", err)
		return
	}
	name = tag.Name
	if tag.Type != Compound {
		err = (os.ErrorString)(fmt.Sprint("nbt.ReadTagCompound: expected compound type, got ", tag.Type))
		return
	}
	if payload, err = reader.ReadCompound(); err != nil {
		error.NewError("could not read compound tag", err)
		return
	}
	return
}

func (reader *TagReader) readPayload(ttype TagType) (payload interface{}, err os.Error) {
	switch ttype {
	case End:
		err = (os.ErrorString)("nbt.readPayload: tag type End has no payload")
	case Byte:
		payload, err = reader.ReadInt8()
		if err != nil {
			err = error.NewError("could not read payload byte", err)
		}
	case Short:
		payload, err = reader.ReadInt16()
		if err != nil {
			err = error.NewError("could not read payload short", err)
		}
	case Int:
		payload, err = reader.ReadInt32()
		if err != nil {
			err = error.NewError("could not read payload int", err)
		}
	case Long:
		payload, err = reader.ReadInt64()
		if err != nil {
			err = error.NewError("could not read payload long", err)
		}
	case Float:
		payload, err = reader.ReadFloat32()
		if err != nil {
			err = error.NewError("could not read payload float", err)
		}
	case Double:
		payload, err = reader.ReadFloat64()
		if err != nil {
			err = error.NewError("could not read payload double", err)
		}
	case ByteArray:
		payload, err = reader.ReadByteArray()
		if err != nil {
			err = error.NewError("could not read payload byte array", err)
		}
	case String:
		payload, err = reader.ReadString()
		if err != nil {
			err = error.NewError("could not read payload string", err)
		}
	case List:
		payload, err = reader.ReadList()
		if err != nil {
			err = error.NewError("could not read payload list", err)
		}
	case Compound:
		payload, err = reader.ReadCompound()
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

func (reader *TagReader) ReadBool() (b bool, err os.Error) {
	var boolSByte int8
	if boolSByte, err = reader.ReadInt8(); err != nil {
		err = error.NewError("could not read boolean value", err)
		return
	}
	b = boolSByte != 0 // FIXME? ensure always 0 or 1?
	return
}

func (writer *TagWriter) WriteBool(b bool) (err os.Error) {
	var boolSByte int8
	if b {
		boolSByte = 1
	} else {
		boolSByte = 0
	}
	if err = writer.WriteInt8(boolSByte); err != nil {
		err = error.NewError("could not write boolean value", err)
		return
	}
	return
}

func (reader *TagReader) ReadByteArray() (b []byte, err os.Error) {
	// Normally, we'd use the reader's buffer to read into.  However, it doesn't
	// buy us much here, because it's essentially just scratch space and we need
	// to return something that won't change after we return it.
	var length int32
	if length, err = reader.ReadInt32(); err != nil {
		err = error.NewError("could not read byte array's length", err)
		return
	}
	if length < 0 {
		err = error.NewError("byte array's length cannot be < 0", nil)
	}
	b = make([]byte, length)
	if _, err = io.ReadFull(reader.r, b); err != nil {
		err = error.NewError("could not read byte array", err)
	}
	return
}

func (reader *TagReader) ReadCompound() (c map[string]interface{}, err os.Error) {
	c = make(map[string]interface{})
	var tag NamedTag
	for {
		if tag, err = reader.ReadNamedTag(); err != nil {
			err = error.NewError("could not read named tag", err)
			return
		}
		if tag.Type == End {
			return
		}
		if c[tag.Name], err = reader.readPayload(tag.Type); err != nil {
			err = error.NewError("could not read payload", err)
			return
		}
	}
	panic("shouldn't get here")
}

func (reader *TagReader) ReadFloat32() (f float32, err os.Error) {
	var i32 int32
	if i32, err = reader.ReadInt32(); err != nil {
		return
	}
	f = math.Float32frombits(uint32(i32))
	return
}

func (writer *TagWriter) WriteFloat32(f float32) (err os.Error) {
	ui32 := math.Float32bits(f)
	if err = writer.WriteInt32(int32(ui32)); err != nil {
		return
	}
	return
}

func (reader *TagReader) ReadFloat64() (f float64, err os.Error) {
	var i64 int64
	if i64, err = reader.ReadInt64(); err != nil {
		return
	}
	f = math.Float64frombits(uint64(i64))
	return
}

func (writer *TagWriter) WriteFloat64(f float64) (err os.Error) {
	ui64 := math.Float64bits(f)
	if err = writer.WriteInt64(int64(ui64)); err != nil {
		return
	}
	return
}

func (reader *TagReader) ReadInt8() (i int8, err os.Error) {
	var bytes = reader.buf[0:1]
	if _, err = io.ReadFull(reader.r, bytes); err != nil {
		return
	}
	i = int8(bytes[0])
	return
}

func (writer *TagWriter) WriteInt8(i int8) (err os.Error) {
	var bytes = writer.buf[0:1]

	bytes[0] = byte(i)

	if _, err = writer.w.Write(bytes); err != nil {
		return
	}

	return
}

func (reader *TagReader) ReadInt16() (i int16, err os.Error) {
	var bytes = reader.buf[0:2]
	if _, err = io.ReadFull(reader.r, bytes); err != nil {
		return
	}
	i = int16(uint16(bytes[1]) | uint16(bytes[0])<<8)
	return
}

func (writer *TagWriter) WriteInt16(i int16) (err os.Error) {
	var bytes = writer.buf[0:2]
	ui := uint16(i)

	bytes[0] = byte(ui >> 8)
	bytes[1] = byte(ui)

	if _, err = writer.w.Write(bytes); err != nil {
		return
	}

	return
}

func (reader *TagReader) ReadInt32() (i int32, err os.Error) {
	var bytes = reader.buf[0:4]
	if _, err = io.ReadFull(reader.r, bytes); err != nil {
		return
	}
	i = int32(uint32(bytes[3]) | uint32(bytes[2])<<8 | uint32(bytes[1])<<16 | uint32(bytes[0])<<24)
	return
}

func (writer *TagWriter) WriteInt32(i int32) (err os.Error) {
	var bytes = writer.buf[0:4]
	ui := uint32(i)

	bytes[0] = byte(ui >> 24)
	bytes[1] = byte(ui >> 16)
	bytes[2] = byte(ui >> 8)
	bytes[3] = byte(ui)

	if _, err = writer.w.Write(bytes); err != nil {
		return
	}

	return
}

func (reader *TagReader) ReadInt64() (i int64, err os.Error) {
	var bytes = reader.buf[0:8]
	if _, err = io.ReadFull(reader.r, bytes); err != nil {
		return
	}
	i = int64(uint64(bytes[7]) | uint64(bytes[6])<<8 | uint64(bytes[5])<<16 | uint64(bytes[4])<<24 | uint64(bytes[3])<<32 | uint64(bytes[2])<<40 | uint64(bytes[1])<<48 | uint64(bytes[0])<<56)
	return
}

func (writer *TagWriter) WriteInt64(i int64) (err os.Error) {
	var bytes = writer.buf[0:8]
	ui := uint64(i)

	bytes[0] = byte(ui >> 56)
	bytes[1] = byte(ui >> 48)
	bytes[2] = byte(ui >> 40)
	bytes[3] = byte(ui >> 32)
	bytes[4] = byte(ui >> 24)
	bytes[5] = byte(ui >> 16)
	bytes[6] = byte(ui >> 8)
	bytes[7] = byte(ui)

	if _, err = writer.w.Write(bytes); err != nil {
		return
	}

	return
}

func (reader *TagReader) ReadList() (l []interface{}, err os.Error) {

	var ttypei8 int8
	var llen int32
	if ttypei8, err = reader.ReadInt8(); err != nil {
		err = error.NewError("could not read list type", err)
		return
	}
	if llen, err = reader.ReadInt32(); err != nil {
		err = error.NewError("could not read list type", err)
		return
	}
	if llen < 0 {
		err = error.NewError("list length cannot be < 0", nil)
		return
	}
	ttype := TagType(ttypei8)
	l = make([]interface{}, int(llen))
	for i := int32(0); i < llen; i++ {
		var payload interface{}
		if payload, err = reader.readPayload(ttype); err != nil {
			err = error.NewError(fmt.Sprint("could not read list payload at index", i), nil)
			return
		}
		l[i]=payload
	}
	return
}

func (reader *TagReader) ReadString() (s string, err os.Error) {
	var strlen int16
	if strlen, err = reader.ReadInt16(); err != nil {
		return
	}
	var strchars []byte
	if int(strlen) <= len(reader.buf) {
		strchars = reader.buf[0:strlen]
	} else {
		strchars = make([]byte, strlen)
	}
	if _, err = io.ReadFull(reader.r, strchars); err != nil {
		return
	}
	s = string(strchars)
	return
}

func (writer *TagWriter) WriteString(s string) (err os.Error) {
	var strlenui int16
	var strlen int

	strlen = len(s)
	if strlen > math.MaxInt16 {
		return (os.ErrorString)("nbt.WriteString: string was too long")
	}
	strlenui = (int16)(strlen)
	if err = writer.WriteInt16(strlenui); err != nil {
		return
	}
	strchars := ([]byte)(s)
	if _, err = writer.w.Write(strchars); err != nil {
		return
	}
	s = string(strchars)
	return
}
