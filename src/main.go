package main

import (
	"bufio"
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net"
	"os"
	"strings"
	"time"
	"watchman/src/config"
	"watchman/src/logger"

	"github.com/fsnotify/fsnotify"
)

type contentUpdate struct {
	FileName      string   `json:"filename"`
	UpdateContent []string `json:"updatecontent"`
}

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
func sendUDP(msg chan contentUpdate, ipaddr string, port int, myLog *zap.Logger) {
	// 发送数据
	for {
		select {
		case incoming := <-msg:
			// 创建连接
			socket, err := net.Dial("udp", ipaddr+":"+cast.ToString(port))
			if err != nil {
				myLog.Error("udp连接失败", zap.String("reason", err.Error()))
			}

			jsonSend, errjson := json.Marshal(incoming)
			if errjson != nil {
				myLog.Error(errjson.Error())
			}
			sendData := jsonSend
			_, err = socket.Write(sendData)
			if err != nil {
				myLog.Error("发送数据失败!", zap.String("reason", err.Error()))
			} else {
				myLog.Info("UDP数据发送成功", zap.String("filename", incoming.FileName), zap.Strings("content", incoming.UpdateContent))
			}

			socket.Close()
		default:
			// fmt.Printf("empty\n")
			time.Sleep(time.Millisecond * 10)
		}
	}
}

//通过配置文件查找当前文件已读取的位置
func findStart(fileName string, watchlist []config.Record) int {
	res := -1
	for i := 0; i < len(watchlist); i++ {
		if fileName == watchlist[i].File {
			//log.Println("find", fileName, watchlist[i].File)
			res = watchlist[i].Column
		}
	}
	return res
}

//更新配置文件中当前文件的读取位置
func updateStart(fileName string, watchlist []config.Record, newCount int) []config.Record {
	for i := 0; i < len(watchlist); i++ {
		if fileName == watchlist[i].File {
			watchlist[i].Column = newCount
		}
	}
	return watchlist
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

	updateMsg := make(chan contentUpdate, 300)

	mylog.Info("WathMan is starting!")

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
					// 判断是否是配置文件中watchall--配置文件为空则监控全部
					if !myConfig.Watchman.WatchAll { //如果不是监控全部
						//判断当前更改文件是否是filelist中的数据
						findflag := false
						for j := 0; j < len(myConfig.Watchman.FileList); j++ {
							if strings.Contains(event.Name, myConfig.Watchman.FileList[j]) {
								findflag = true
							}
						}
						if findflag == false { //如果没有找到，则本次监控跳过
							mylog.Info("This file is not in watchlist, skip...")
							continue
						}
					}
					// 重新读取配置文件
					myConfig := config.GetConfig()
					// 查看记录中是否存在读取位置
					resFind := findStart(event.Name, myConfig.Records)
					//mylog.Info("FIND*********" + cast.ToString(resFind))
					if resFind != -1 {
						readStart = resFind
					}

					// 读取文件内容
					res, count := readLine(event.Name, readStart)
					mylog.Info(event.Name, zap.Strings("content", res))
					tempRecords := myConfig.Records
					tempRes := config.Record{File: event.Name, Column: count}

					if resFind == -1 { //读取不到->插入记录
						tempRecords = append(tempRecords, tempRes)
					} else { // 能读取到->更新记录
						tempRecords = updateStart(event.Name, myConfig.Records, count)
					}
					// 更新配置文件
					viper.Set("records", tempRecords)

					if config.UpdateConfig() {
						// mylog.Info("Config update successful")
					} else {
						mylog.Error("Config update failed")
					}

					if len(res) != 0 {
						//发送到管道
						tempCotent := contentUpdate{FileName: event.Name, UpdateContent: res}
						updateMsg <- tempCotent
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

	go sendUDP(updateMsg, myConfig.Udpinfo.Host, myConfig.Udpinfo.Port, mylog)

	<-done
}
