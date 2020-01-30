// LianDi - 链滴笔记，链接点滴
// Copyright (c) 2020-present, b3log.org
//
// Lute is licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.

package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/88250/liandi/command"
	"github.com/88250/liandi/util"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	util.InitLog()
	util.InitConf()
	util.InitMount()
	util.InitSearch()

	go util.ParentExited()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	m := melody.New()
	m.Config.MaxMessageSize = 1024 * 1024 * 2
	r.GET("/ws", func(c *gin.Context) {
		if err := m.HandleRequest(c.Writer, c.Request); nil != err {
			util.Logger.Errorf("处理命令失败：%s", err)
		}
	})

	m.HandleConnect(func(s *melody.Session) {
		util.SetPushChan(s)
		util.Logger.Debug("websocket connected")
	})

	m.HandleDisconnect(func(s *melody.Session) {
		util.Logger.Debugf("websocket disconnected")
	})

	m.HandleError(func(s *melody.Session, err error) {
		util.Logger.Debugf("websocket on error: %s", err)
	})

	m.HandleClose(func(s *melody.Session, i int, str string) error {
		util.Logger.Debugf("websocket on close: %v, %v", i, str)
		return nil
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		util.Logger.Debugf("request [%s]", msg)
		request := map[string]interface{}{}
		if err := json.Unmarshal(msg, &request); nil != err {
			result := util.NewResult()
			result.Code = -1
			result.Msg = "Bad Request"
			responseData, _ := json.Marshal(result)
			util.Push(responseData)
			return
		}

		cmdStr := request["cmd"].(string)
		cmdId := request["reqId"].(float64)
		param := request["param"].(map[string]interface{})
		cmd := command.NewCommand(cmdStr, cmdId, param)
		if nil == cmd {
			result := util.NewResult()
			result.Code = -1
			result.Msg = "查找命令 [" + cmdStr + "] 失败"
			util.Push(result.Bytes())
			return
		}
		command.Exec(cmd)
	})

	handleSignal()

	addr := "127.0.0.1:" + util.ServerPort
	util.Logger.Infof("内核进程 [v%s] 正在启动，监听端口 [%s]", util.Ver, "http://"+addr)
	if err := r.Run(addr); nil != err {
		util.Logger.Errorf("启动链滴笔记内核失败 [%s]", err)
	}
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		s := <-c
		util.Logger.Infof("收到系统信号 [%s]，退出内核进程", s)

		util.Close()
		os.Exit(0)
	}()
}
