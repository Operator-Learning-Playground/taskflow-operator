package cmd

var (
	waitFile        string // 等待此文件存在才能运行
	out             string // 执行结果输出到哪  默认输出到当前目录下的out文件
	command         string // 入口
	args            string // 参数
	waitFileContent string // 当等待文件存在时 ，还要同时判断内容
	quitContent     string // 如果waitFileContent设置了值， 当waitFile内容==quitContent时，则退出程序
	encodeFile      string // 这代表是加密文件 ,有这个参数 则无视 command 和args
)
