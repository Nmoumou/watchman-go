# **watchman-go**
golang Version of watchman, which can monitor the text content in the folder and send it through UDP or MQTT

golang版本的watchman，可以监听文件夹中的文本内容，通过UDP或MQTT发送

配置文件

~~~yaml
loginfo:
  loglevel: debug     #log等级 
                      #debug 可以打印出 info debug warn 
                      #info  级别可以打印 warn info
                      #warn  只能打印 warn
                      #debug->info->warn->error
  logpath: ./watchman.log  #log保存路径
  maxage: 15              #保存天数
  maxsize: 20             #单个日志文件大小MB
  servicename: mabowatchman  #监控服务名称
mqttinfo:
  host: 127.0.0.1   #MQTT服务器地址
  password: aaadkeifkd #MQTT连接密码
  port: 1883     #MQTT连接端口
  pubtopic: test  #MQTT发布主题
  qos: 0          #QOS等级
  username: watchman01 #MQTT连接用户名
records:  #监控记录,保存已监控文本名称及当前记录行数
  - file: D:\GoPrj\abc\abc.txt
    column: 32
  - file: add.txt
    column: 1
  - file: D:\GoPrj\abc\aaane.txt
    column: 2
udpinfo: #UDP连接信息
  host: 127.0.0.1
  port: 60000
watchman: #程序配置
  filelist: #监控列表，如果watchall为false,则从此列表中匹配要监控的文件
    - abc.txt
    - efg.txt
  path: 
    - D:\\GoPrj\\abc  #监控文件夹地址(文件夹下新创建的文件夹会自动添加)
  startcolumn: 0      #监控文件起始行数
  suffix: #配置要监控的文件后缀
    - txt
    - log
  transfermethod: udp   #发送方式 (1)upd (2)mqtt (3)both
  watchall: true        #是否监控文件夹内所有文件

~~~

UDP输出格式：

~~~json
{"filename":"D:\\abc.txt","updatecontent":["one column update","two column update"]}
~~~

filename 更新文件路径及名称

updatecontent 字符串列表,每项表示一行更新内容
