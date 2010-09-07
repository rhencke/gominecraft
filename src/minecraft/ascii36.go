package world

// encoding/decoding for minecraft-style base36
var b36chars = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'a', 'b', 'c', 'd', 'e', 'f',
	'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
	'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
	'w', 'x', 'y', 'z',
}

func int32ToBase36String(i int32) string {
	var str [7]byte // 6 possible digits + 1 for sign
	var ix = 7
	var wasneg bool
	if i < 0 {
		wasneg = true
		i = -i
	}
	for {
		ix--
		rem := i % 36
		i = i / 36
		str[ix] = b36chars[rem]
		if i == 0 {
			break
		}
	}
	if wasneg {
		ix--
		str[ix] = '-'
	}
	return string(str[ix:])
}
