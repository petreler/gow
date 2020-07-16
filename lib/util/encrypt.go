package util

import (
	"crypto/md5"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"strings"
)

//GetUUID UUID MD5
func GetUUID() (string, error) {
	u := uuid.NewV4()
	m := md5.New()
	io.WriteString(m, u.String())
	str := strings.ToLower(fmt.Sprintf("%x", m.Sum(nil)))
	if len(str) <= 16 {
		return str, nil
	}
	return str[:16], nil
}

//MD5 编码
//32位长度的小写md5输出
func MD5(str string) string {
	m := md5.New()
	io.WriteString(m, str)
	return strings.ToLower(fmt.Sprintf("%x", m.Sum(nil)))
}
