package qtff

import "encoding/binary"

type FourCC uint32

func (t FourCC) String() string {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(t))
	return string(buf[:])
}

func FourCCFromString(s string) FourCC {
	return FourCC(binary.BigEndian.Uint32([]byte(s)))
}
