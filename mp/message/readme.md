### 被动接收消息和回复示例
```Go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/message/request"
	"github.com/chanxuehong/wechat2/mp/message/response"
	"github.com/chanxuehong/wechat2/util"
)

// 处理普通文本消息, 原样返回
func TextMessageHandler(w http.ResponseWriter, r *mp.Request) {
	textReq := request.GetText(r.MixedMsg)
	textResp := response.NewText(textReq.FromUserName, textReq.ToUserName,
		textReq.Content, textReq.CreateTime)

	if err := mp.WriteAESResponse(w, r, textResp); err != nil {
		log.Println(err)
	}
}

// 上报地理位置事件处理
func LocationEventHandler(w http.ResponseWriter, r *mp.Request) {
	event := request.GetLocationEvent(r.MixedMsg)
	fmt.Println(event) // 处理事件
}

func main() {
	aesKey, err := util.AESKeyDecode("encodedAESKey")
	if err != nil {
		panic(err)
	}

	messageServeMux := mp.NewMessageServeMux()
	messageServeMux.MessageHandleFunc(request.MsgTypeText, TextMessageHandler)
	messageServeMux.EventHandleFunc(request.EventTypeLocation, LocationEventHandler)

	wechatServer := mp.NewDefaultWechatServer("id", "token", "appid", messageServeMux, aesKey)

	wechatServerFrontend := mp.NewWechatServerFrontend(wechatServer, nil)

	http.Handle("/wechat", wechatServerFrontend)
	http.ListenAndServe(":80", nil)
}
```