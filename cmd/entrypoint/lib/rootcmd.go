package lib

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "entrypoint",
	Short: "cicd operator EntryPoint",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		CheckFlags()    //检查参数合法性
		CheckWaitFile() //检查 等待文件是否存在
		//业务逻辑
		fmt.Println("业务逻辑")
		ExecCmdAndArgs(args)
	},
}

// InitCmd 初始化
func InitCmd() {
	rootCmd.Flags().StringVar(&waitFile, "wait", "", "entrypoint -wait /var/run/1")
	rootCmd.Flags().StringVar(&out, "out", "", "entrypoint -out /var/run/out")
	rootCmd.Flags().StringVar(&command, "command", "", "entrypoint -command bash")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

/*
	go run cmd/main.go --wait ./1.txt --out out.log --command go version
*/
