package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

// 测试方式
//  go run cmd/container-agent/main.go --wait ./1.txt --out out.log --command go version
// ./container-agent  --out stdout --wait 1.txt --command="sh -c" "echo 123"
// ./container-agent  --encodefile /xxx/xxx/1.txt
var rootCmd = &cobra.Command{
	Use:   "container-agent",
	Short: "container agent",
	Run: func(cmd *cobra.Command, args []string) {

		ValidateFlags() // 检查参数合法性
		CheckWaitFile() // 检查等待文件是否存在
		ExecCmdAndArgs(args)

	},
}

// InitCmd 初始化
func InitCmd() {
	rootCmd.Flags().StringVar(&waitFile, "wait", "", "container-agent --wait /var/run/1")
	// 增加了一个参数。 如果有这个参数，那么还得判断 内容是否匹配
	// 如果没有这个参数，则只判断是否 有 wait 对应的文件
	rootCmd.Flags().StringVar(&waitFileContent, "waitcontent", "", "container-agent --wait /var/run/1 --waitcontent 2 ")
	rootCmd.Flags().StringVar(&out, "out", "", "container-agent --out /var/run/out")
	rootCmd.Flags().StringVar(&command, "command", "", "container-agent --command bash")
	rootCmd.Flags().StringVar(&quitContent, "quit", "-1", "container-agent --quit -2")
	rootCmd.Flags().StringVar(&encodeFile, "encodefile", "", "container-agent --encodefile /var/run/abc")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
