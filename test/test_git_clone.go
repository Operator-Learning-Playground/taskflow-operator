package main

import (
	"github.com/go-git/go-git/v5"
	"log"
	"os"
)

/*
	获取git仓库代码的调研
*/

func main() {
	dir := "./test/git"
	os.RemoveAll(dir)
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: "https://github.com/googs1025/Golang-Design-Pattern-Demo",
		// 用来获取不同分支或tag
		//ReferenceName: "refs/heads/test", //refs/tags/xxxoo
	})
	if err != nil {
		log.Fatalln(err)
	}

}
