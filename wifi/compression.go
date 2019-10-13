package wifi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
)

func Compress(data []byte) (bool, []byte, error) {
	oldSize := len(data)
	buf := bytes.Buffer{}
	if zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression); err != nil {
		return false, nil, fmt.Errorf("error initializing payload compression: %v", err)
	} else if _, err := zw.Write(data); err != nil {
		return false, nil, fmt.Errorf("error during payload compression: %v", err)
	} else if err = zw.Close(); err != nil {
		return false, nil, fmt.Errorf("error while finalizing payload compression: %v", err)
	}

	compressed := buf.Bytes()
	newSize := len(compressed)

	// log.Debug("gzip: %d > %d", oldSize, newSize)

	if newSize < oldSize {
		return true, compressed, nil
	}
	return false, data, nil
}

func Decompress(data []byte) ([]byte, error) {
	if zr, err := gzip.NewReader(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("error initializing payload decompression: %v", err)
	} else {
		defer zr.Close()
		return ioutil.ReadAll(zr)
	}
}
