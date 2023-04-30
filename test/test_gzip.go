package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
)

/*
	用来处理脚本中特殊符号的问题
*/

// Gzip 将脚本字符串压缩转换成base64，避免
func Gzip(s string) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(s))
	if err != nil {
		klog.Error(err)
		return ""
	}

	err = gz.Close()
	if err != nil {
		klog.Error(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())

}

// UnGzip 解压缩
func UnGzip(s string) string {
	dbyte, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		klog.Error(err)
		return ""
	}
	read_data := bytes.NewReader(dbyte)
	reader, err := gzip.NewReader(read_data)
	if err != nil {
		klog.Error(err)
		return ""
	}
	defer reader.Close()

	res, err := ioutil.ReadAll(reader)
	if err != nil {
		klog.Error(err)
		return ""
	}

	return string(res)

}

func main() {
	t := Gzip("echo 111111")
	fmt.Println(t)

	fmt.Println(UnGzip(t))
}
