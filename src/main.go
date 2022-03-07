package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"watchman/src/config"
	"watchman/src/logger"

	"github.com/fsnotify/fsnotify"
)

//从指定行开始读取文本内容
func readLine(fileName string, lineNumber int) ([]string, int) {
	file, _ := os.Open(fileName)
	fileScanner := bufio.NewScanner(file)
	lineCount := 1
	var res []string
	for fileScanner.Scan() {
		if lineCount >= lineNumber {
			scanRes := fileScanner.Text()
			if len(scanRes) > 0 {
				res = append(res, scanRes)
			}
		}
		lineCount++
	}
	defer file.Close()

	return res, lineCount
}

//通过UDP发送数据
func sendUDP(msg string) int {
	// 创建连接
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IP("127.0.0.1"),
		Port: 60000,
	})
	if err != nil {
		fmt.Println("连接失败!", err)
		return -1
	}
	defer socket.Close()

	// 发送数据
	senddata := []byte(msg)
	_, err = socket.Write(senddata)
	if err != nil {
		fmt.Println("发送数据失败!", err)
		return -1
	}
	return 0
}

//通过配置文件查找当前文件已读取的位置
func findStart(fileName string, watchlist []config.Record) int {
	res := -1
	for i := 0; i < len(watchlist); i++ {
		if fileName == watchlist[i].File {
			log.Println("find", fileName, watchlist[i].File)
			res = watchlist[i].Column
		}
	}
	return res
}

func main() {

	myConfig := config.GetConfig()
	// 从配置文件中初始化参数
	mylog := logger.InitLogger(myConfig.Loginfo.LogPath,
		myConfig.Loginfo.LogLevel,
		myConfig.Loginfo.MaxSize,
		myConfig.Loginfo.MaxAge,
		myConfig.Loginfo.ServiceName)

	readStart := myConfig.Watchman.StartColumn

	mylog.Info("Wathman is starting!")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		mylog.Error(err.Error())
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// logger.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					mylog.Info("modified file:" + event.Name)
					// 查看记录中是否存在读取位置
					resFind := findStart(event.Name, myConfig.Records)
					mylog.Info("FIND*********" + cast.ToString(resFind))
					if resFind != -1 {
						readStart = resFind
					}

					// 读取文件内容
					res, count := readLine(event.Name, readStart)
					mylog.Info(cast.ToString(res))
					mylog.Info(cast.ToString(count))

					// 能读取到->更新记录,读取不到->插入记录
					viper.Set("udpinfo.port", 555555)
					if config.UpdateConfig() {
						mylog.Info("Config update successful")
					}
					if len(res) != 0 {
						readStart++
						//发送UDP
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				mylog.Error(err.Error())
			}
		}
	}()

	err = watcher.Add(myConfig.Watchman.Path)
	if err != nil {
		mylog.Error(err.Error())
	}
	<-done
}
