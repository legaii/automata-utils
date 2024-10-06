package ioutils

import (
	"fmt"
	"io"
)

func Read[ValueType any](reader io.Reader) ValueType {
	var value ValueType
	if _, err := fmt.Fscan(reader, &value); err != nil {
		panic(err)
	}
	return value
}

func Write[ValueType any](writer io.Writer, value ValueType) {
	if _, err := fmt.Fprint(writer, value); err != nil {
		panic(err)
	}
}
