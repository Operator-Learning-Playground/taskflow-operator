package lib

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"time"
)

// CheckFlags 校验参数
func CheckFlags() {
	if waitFile == "" || out == "" {
		log.Println("错误的参数")
		os.Exit(1)
	}
}

// CheckWaitFile 检查等待文件是否存在
func CheckWaitFile() {
	for {
		if _, err := os.Stat(waitFile); err == nil {
			return
		} else if errors.Is(err, os.ErrNotExist) {
			// 文件不存在，继续等待
			time.Sleep(time.Millisecond * 20)
			continue
		} else {
			log.Fatal(err)
		}
	}
}

// ExecCmdAndArgs 程序运行入口和参数
func ExecCmdAndArgs(args []string) {
	logF, err := os.OpenFile(out, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0655)
	if err != nil {
		log.Fatal(err)
	}

	defer logF.Close()
	exc := exec.Command(command, args...) // 选择command命令
	// 执行结果从放入文件
	exc.Stdout = logF
	exc.Stderr = logF
	if err = exc.Run(); err != nil {
		log.Fatal(err)
	}

}
