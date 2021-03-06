// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

// 回调 URL 上索引 WechatServer 的 key 的名称.
//  比如下面的回调地址里面就可以根据 wechat1 来索引对应的 WechatServer.
//  http://www.xxx.com/?wechatkey=wechat1&signature=XXX&timestamp=123456789&nonce=12345678
const URLQueryWechatKeyName = "wechatkey"

// 多个 WechatServer 的前端, 负责处理 http 请求, net/http.Handler 的实现
//
//  NOTE:
//  MultiWechatServerFrontend 可以处理多个公众号的消息（事件），但是要求在回调 URL 上加上一个
//  查询参数，参考常量 URLQueryWechatKeyName，这个参数的值就是 MultiWechatServerFrontend
//  索引 WechatServer 的 key。
//
//  例如回调 URL 为 http://www.xxx.com/weixin?wechatkey=1234567890，那么就可以在后端调用
//
//    MultiWechatServerFrontend.SetWechatServer("1234567890", WechatServer)
//
//  来增加一个 WechatServer 来处理 wechatkey=1234567890 的消息（事件）。
//
//  MultiWechatServerFrontend 并发安全，可以在运行中动态增加和删除 WechatServer。
type MultiWechatServerFrontend struct {
	rwmutex               sync.RWMutex
	wechatServerMap       map[string]WechatServer
	invalidRequestHandler InvalidRequestHandler
}

// 设置 InvalidRequestHandler, 如果 handler == nil 则使用默认的 DefaultInvalidRequestHandler
func (front *MultiWechatServerFrontend) SetInvalidRequestHandler(handler InvalidRequestHandler) {
	front.rwmutex.Lock()
	defer front.rwmutex.Unlock()

	if handler == nil {
		front.invalidRequestHandler = DefaultInvalidRequestHandler
	} else {
		front.invalidRequestHandler = handler
	}
}

// 设置 wechatkey-WechatServer pair.
// 如果 wechatkey == "" 或者 wechatServer == nil 则不做任何操作
func (front *MultiWechatServerFrontend) SetWechatServer(wechatkey string, wechatServer WechatServer) {
	if wechatkey == "" {
		return
	}
	if wechatServer == nil {
		return
	}

	front.rwmutex.Lock()
	defer front.rwmutex.Unlock()

	if front.wechatServerMap == nil {
		front.wechatServerMap = make(map[string]WechatServer)
	}
	front.wechatServerMap[wechatkey] = wechatServer
}

// 删除 wechatkey 对应的 WechatServer
func (front *MultiWechatServerFrontend) DeleteWechatServer(wechatkey string) {
	front.rwmutex.Lock()
	defer front.rwmutex.Unlock()

	delete(front.wechatServerMap, wechatkey)
}

// 删除所有的 WechatServer
func (front *MultiWechatServerFrontend) DeleteAllWechatServer() {
	front.rwmutex.Lock()
	defer front.rwmutex.Unlock()

	front.wechatServerMap = make(map[string]WechatServer)
}

// 实现 http.Handler
func (front *MultiWechatServerFrontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		front.rwmutex.RLock()
		invalidRequestHandler := front.invalidRequestHandler
		front.rwmutex.RUnlock()

		if invalidRequestHandler == nil {
			invalidRequestHandler = DefaultInvalidRequestHandler
		}
		invalidRequestHandler.ServeInvalidRequest(w, r, err)
		return
	}

	wechatKey := urlValues.Get(URLQueryWechatKeyName)
	if wechatKey == "" {
		front.rwmutex.RLock()
		invalidRequestHandler := front.invalidRequestHandler
		front.rwmutex.RUnlock()

		if invalidRequestHandler == nil {
			invalidRequestHandler = DefaultInvalidRequestHandler
		}
		err = fmt.Errorf("the url query value with name %s is empty", URLQueryWechatKeyName)
		invalidRequestHandler.ServeInvalidRequest(w, r, err)
		return
	}

	front.rwmutex.RLock()
	invalidRequestHandler := front.invalidRequestHandler
	wechatServer := front.wechatServerMap[wechatKey]
	front.rwmutex.RUnlock()

	if invalidRequestHandler == nil {
		invalidRequestHandler = DefaultInvalidRequestHandler
	}
	if wechatServer == nil {
		invalidRequestHandler.ServeInvalidRequest(w, r, fmt.Errorf("Not found WechatServer for %s == %s", URLQueryWechatKeyName, wechatKey))
		return
	}

	ServeHTTP(w, r, urlValues, wechatServer, invalidRequestHandler)
}
