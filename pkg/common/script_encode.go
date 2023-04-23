package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"log"
)

// EncodeScript 加密脚本
func EncodeScript(str string) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write([]byte(str))
	if err != nil {
		log.Println(err)
		return ""
	}

	// 需要关掉
	err = gz.Close()
	if err != nil {
		log.Println(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
