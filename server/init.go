package server

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/micro-kit/micro-common/common"
)

// 初始化.env文件
func init() {
	envFile := common.GetRootDir() + ".env"
	if ext, _ := common.PathExists(envFile); ext == true {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Println("读取.env错误", err)
		}
	}
}
