package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestINI_GetKey(t *testing.T) {
	fmt.Println(DefaultString("app_name", "gow"))
	fmt.Println(DefaultString("app_mode", "dev"))
	fmt.Println(DefaultString("http_port", "8080"))

	fmt.Println(DefaultString("gkzy-user::user", "zituocn"))

	strs := Keys("gkzy-user")
	fmt.Println(strings.Join(strs, ","))
}
