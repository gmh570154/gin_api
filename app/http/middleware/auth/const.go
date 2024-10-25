/*基于ak, sk实现的服务认证中间件
 */
package auth

import (
	"hash"
)

// HashFunc 返回一个hash.Hash接口
type HashFunc func() hash.Hash

// KeyFunc 查询accesskey,返回secretKey的函数
type KeyFunc func(accessKey string) (secretKey string)

type store map[string]string

var keyStore = store{"access_key": "secret_key"}

// GetKeyFunc 返回aksk.KeyFunc
func GetKeyFunc() KeyFunc {
	return func(ak string) string {
		return keyStore[ak]
	}
}

const (
	// headerAccessKey 访问key
	authorization   = `Authorization`
	headerAccessKey = `Credential`
	// headerSignature 签名hmac的签名
	headerSignature = `Signature`
	// headerBodyHash http的Body hash计算的mac
	headerBodyHash = `X-Acs-Content-Sha256`
	// headerRandomStr 随机字符串
	headerRandomStr = `X-Acs-Signature-Nonce`
	signedHeaders   = `SignedHeaders`
)
