package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

// 测试方式
//  go run cmd/entrypoint/main.go --wait ./1.txt --out out.log --command go version
//./entrypoint  --out stdout --wait 1.txt --command="sh -c" "echo 123"
//./entrypoint  --encodefile /home/shenyi/mycicd/ep/1.txt
var rootCmd = &cobra.Command{
	Use:   "entrypoint",
	Short: "container entrypoint",
	Run: func(cmd *cobra.Command, args []string) {

		CheckFlags()    //检查参数合法性
		CheckWaitFile() //检查 等待文件是否存在
		// 业务逻辑
		ExecCmdAndArgs(args)

	},
}

// InitCmd 初始化
func InitCmd() {
	rootCmd.Flags().StringVar(&waitFile, "wait", "", "entrypoint --wait /var/run/1")
	// 增加了一个参数。 如果有这个参数，那么还得判断 内容是否匹配
	// 如果没有这个参数，则只判断是否 有 wait 对应的文件
	rootCmd.Flags().StringVar(&waitFileContent, "waitcontent", "", "entrypoint --wait /var/run/1 --waitcontent 2 ")
	rootCmd.Flags().StringVar(&out, "out", "", "entrypoint --out /var/run/out")
	rootCmd.Flags().StringVar(&command, "command", "", "entrypoint --command bash")
	rootCmd.Flags().StringVar(&quitContent, "quit", "-1", "entrypoint --quit -2")
	rootCmd.Flags().StringVar(&encodeFile, "encodefile", "", "entrypoint --encodefile /var/run/abc")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
