package xen

import (
	"bytes"

	"github.com/radhus/xenmatchbox/machine"
)

// OutputSimple0 outputs with simple0 syntax
func OutputSimple0(conf *machine.Configuration) []byte {
	var buf bytes.Buffer

	buf.WriteString("kernel ")
	buf.WriteString(conf.Kernel)
	buf.WriteByte(0x00)

	buf.WriteString("ramdisk ")
	buf.WriteString(conf.Ramdisk)
	buf.WriteByte(0x00)

	buf.WriteString("args ")
	buf.WriteString(conf.Args)
	buf.WriteByte(0x00)

	return buf.Bytes()
}
