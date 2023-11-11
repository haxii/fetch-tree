package fetch_tree

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

type NodeInfo struct {
	ID, Name string
	IsLeaf   bool
}

type Tree map[string]NodeMap
type NodeMap map[string]string

func (c *Tree) DecodeFrom(data []byte) error {
	dataReader := bytes.NewReader(data)
	gzReader, err := gzip.NewReader(dataReader)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	decoder := json.NewDecoder(gzReader)
	err = decoder.Decode(c)
	return err
}

func (c *Tree) EncodeToBytes() ([]byte, error) {
	b := new(bytes.Buffer)
	if err := c.Encode(b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c *Tree) EncodeToByteVar() ([]byte, error) {
	b := new(bytes.Buffer)
	w := newByteVarWriter(b)
	if err := c.Encode(w); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c *Tree) Encode(w io.Writer) error {
	gzWriter := gzip.NewWriter(w)
	encoder := json.NewEncoder(gzWriter)
	err := encoder.Encode(c)
	if err = gzWriter.Flush(); err != nil {
		return err
	}
	return gzWriter.Close()
}
