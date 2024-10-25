package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理函数
type ErrorHandler func(c *gin.Context, err error)

func handleError(c *gin.Context, err error) {
	e, ok := err.(*Error)
	if !ok {
		e = &Error{Message: err.Error()}
	}
	log.Printf("验证请求错误: %s", err)
	c.AbortWithStatusJSON(http.StatusUnauthorized, e)
}

// Validate 返回一个验证请求的gin中间件, keyFn指定了查询SecretKey的函数,如果等于nil,将panic;
func Validate() gin.HandlerFunc {
	log.Printf("启用aksk认证")
	return func(c *gin.Context) {
		if err := validRequest(c); err != nil {
			handleError(c, err)
			if !c.IsAborted() {
				c.Abort()
			}
		}
	}
}

func validRequest(c *gin.Context) error {
	canonicalQueryString := ""
	keys := maps.Keys(c.Request.URL.Query())
	sort.Strings(keys)
	for _, k := range keys {
		v := c.Request.URL.Query().Get(k)
		canonicalQueryString += percentCode(url.QueryEscape(k)) + "=" + percentCode(url.QueryEscape(v)) + "&"
	}
	canonicalQueryString = strings.TrimSuffix(canonicalQueryString, "&")
	fmt.Printf("canonicalQueryString========>%s\n", canonicalQueryString) // 请求参数排序之后的字符串

	headers := make(map[string]string)
	if authorization_str := c.GetHeader(authorization); authorization_str == "" {
		return ErrAuthorizationEmpty
	} else {
		auth_content := authorization_str[17:] // 截取后面的有效字符串
		auth_arr := strings.Split(auth_content, ",")

		for _, val := range auth_arr {
			item := strings.Split(val, "=")
			headers[item[0]] = item[1]
		}
	}

	var HeadersKeys []string
	if signed_headers, exit := headers[signedHeaders]; !exit {
		return ErrAccessKeyEmpty
	} else {
		HeadersKeys = strings.Split(signed_headers, ";")
		sort.Strings(HeadersKeys) // header key 排序
	}
	hashedRequestPayload := c.GetHeader(headerBodyHash)

	var canonicalHeaders, signedHeaders string

	// fmt.Println(HeadersKeys)
	for _, k := range HeadersKeys {
		lowerKey := strings.ToLower(k)
		if lowerKey == "host" || strings.HasPrefix(lowerKey, "x-acs-") || lowerKey == "content-type" {
			var value string
			if lowerKey == "host" {
				value = c.Request.Host
			} else {
				value = c.GetHeader(k)
			}
			canonicalHeaders += lowerKey + ":" + value + "\n"
			signedHeaders += lowerKey + ";"
		}
	}
	signedHeaders = strings.TrimSuffix(signedHeaders, ";") // 去掉最后的；字符串
	fmt.Println("signedHeaders:", signedHeaders)

	// 请求合并字符串
	canonicalRequest := c.Request.Method + "\n" + c.Request.URL.Path + "\n" + canonicalQueryString + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + hashedRequestPayload
	fmt.Printf("canonicalRequest========>\n%s\n", canonicalRequest)

	// 获取ak/sk
	var sk string
	if ak, exit := headers[headerAccessKey]; !exit {
		return ErrAccessKeyEmpty
	} else {
		if sk = GetKeyFunc()(ak); sk == "" {
			return ErrSecretKeyEmpty
		}
	}

	hashedCanonicalRequest := sha256Hex(canonicalRequest) // hex算法

	var ALGORITHM = "ACS3-HMAC-SHA256"
	stringToSign := ALGORITHM + "\n" + hashedCanonicalRequest
	fmt.Printf("stringToSign========>\n%s\n", stringToSign)

	byteData, err := hmac256([]byte(sk), stringToSign) // 使用sk进行加密 转字节类型
	if err != nil {
		fmt.Println(err)
	}
	signature1 := strings.ToLower(hex.EncodeToString(byteData)) // 16位编码

	if signature, exit := headers[headerSignature]; !exit {
		return ErrSignatueEmpty
	} else {
		if signature1 != signature { // 判断签名是否一致
			return ErrSignatureInvalid
		}
	}

	return nil
}

func hmac256(key []byte, toSignString string) ([]byte, error) {
	// 实例化HMAC-SHA256哈希
	h := hmac.New(sha256.New, key)
	// 写入待签名的字符串
	_, err := h.Write([]byte(toSignString))
	if err != nil {
		return nil, err
	}
	// 计算签名并返回
	return h.Sum(nil), nil
}

func sha256Hex(str string) string {
	// 实例化SHA-256哈希函数
	hash := sha256.New()
	// 将字符串写入哈希函数
	_, _ = hash.Write([]byte(str))
	// 计算SHA-256哈希值并转换为小写的十六进制字符串
	hexString := hex.EncodeToString(hash.Sum(nil))

	return hexString
}

func percentCode(str string) string {
	// 替换特定的编码字符
	str = strings.ReplaceAll(str, "+", "%20")
	str = strings.ReplaceAll(str, "*", "%2A")
	str = strings.ReplaceAll(str, "%7E", "~")
	return str
}
