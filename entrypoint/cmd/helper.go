package cmd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CheckFlags 检查参数
func CheckFlags() {
	if encodeFile == "" {
		if waitFile == "" || out == "" || command == "" {
			log.Println("error input param...")
			os.Exit(1)
		}
	}
}

// GenEncodeFile 生成 waitfile 中的内容
func GenEncodeFile(encodefile string) error {
	f, err := os.OpenFile(encodefile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	decode := UnGzip(string(b)) //反解 之后的 字符串 ,重新写入
	err = f.Truncate(0)         //清空文件1
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0) //清空文件2
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(decode))
	if err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

// getWaitContent 读取waitfile内容
func getWaitContent() string {
	f, err := os.Open(waitFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	ret := string(b)
	return strings.Trim(ret, " ")
}

// CheckWaitFile 检查等待文件是否存在
func CheckWaitFile() {
	for {
		if _, err := os.Stat(waitFile); err == nil {
			// 代表WaitFile存在
			// 此时要判断WaitFileContent 是否设置了内容，如果设置了要判断
			if waitFileContent == "" { //代表没有设置 waitFileContent ，则直接过
				return
			} else {
				getContent := getWaitContent()
				if waitFileContent == getContent { //目标一样，则通过
					return
				}
				// 此时程序要退出
				if getContent == quitContent {
					log.Println("任务被取消")
					os.Exit(1) //停止程序
				}
			}
			// 文件真的不存在
		} else if errors.Is(err, os.ErrNotExist) {
			time.Sleep(time.Millisecond * 500)
			continue
		} else {
			// 其他未知错误
			log.Fatal(err)
		}
	}

}

// ExecCmdAndArgs 运行入口和参数
func ExecCmdAndArgs(args []string) {
	var logF *os.File
	// 判断使用标准输出模式或是日志模式
	if out == "stdout" || out == "" {
		logF = os.Stdout
	} else {
		lf, err := os.OpenFile(out, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0655)
		if err != nil {
			log.Fatal(err)
		}
		logF = lf
		defer logF.Close()
	}

	argList := []string{} //默认是空 字符串切片
	cmd := ""
	// 非加密文件模式
	if encodeFile == "" {
		cmdList := strings.Split(command, " ") // 譬如 sh -c 这种字符串，要切割,取出第一个作为command
		cmd = cmdList[0]                       //肯定有值 ，所以不用判断
		if len(cmdList) > 1 {                  //把剩余的合并到argList 前面
			argList = append(argList, cmdList[1:]...)
		}
		argList = append(argList, args...)

	} else {
		// 加密文件模式
		cmd = "sh" //写死
		argList = []string{encodeFile}
		err := GenEncodeFile(encodeFile) // 反解文件内容 并重新写入
		if err != nil {
			log.Fatal("EncodeFile error:", err)
		}
	}
	exc := exec.Command(cmd, argList...)
	exc.Stdout = logF
	exc.Stderr = logF
	if err := exc.Run(); err != nil {
		log.Fatal("Exec error:", err)
	}

}

func Gzip(str string) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(str))
	if err != nil {
		log.Println(err)
		return ""
	}
	err = gz.Close() //这里要关掉，否则取不到数据  也可手工flush.但依然要关掉gz
	if err != nil {
		log.Println(err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func UnGzip(str string) string {
	dbytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.Println(err)
		return ""
	}
	readData := bytes.NewReader(dbytes)
	reader, err := gzip.NewReader(readData)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer reader.Close()
	ret, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println("read gzip error:", err)
		return ""
	}

	return string(ret)
}
