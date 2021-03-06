// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

// access_token 中控服务器接口, see token_server.png.
type TokenServer interface {
	// 从中控服务器获取 access_token, 该 access_token 一般缓存在某个地方.
	Token() (token string, err error)

	// 请求 access_token 中控服务器到微信服务器刷新 access_token.
	TokenRefresh() (token string, err error)
}
