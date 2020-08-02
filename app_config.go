package gow

import (
	"github.com/gkzy/gow/lib/config"
	"os"
)

const (
	defaultConfig     = "conf/app.conf"
	defaultDevConfig  = "conf/dev.app.conf"
	defaultProdConfig = "conf/prod.app.conf"
)

var (
	fileName string
)

// AppConfig gow 配置入口
// 也可以通过此配置文件统一设置app的基础配置
// 除此之外，也可以通过
//	r := gow.Default()
//	r.AutoRender = true
//  r.AppName = "gow"
//  r.SetView("view") 等方式设置
type AppConfig struct {
	AppName       string //名称
	RunMode       string //运行模式:dev|prod
	HttpAddr      string //"127.0.0.1:8080" or ":8080"
	AutoRender    bool   //是否自动渲染html模板
	Views         string //html 模板目录
	TemplateLeft  string //html模板左符号
	TemplateRight string //html模板右符号
}

// GetAppConfig 获取配置文件中的信息
// 默认使用conf/app.conf配置文件
//  当环境变量 APP_RUN_MODE ="prod"时，使用 conf/prod.app.conf
//  当环境变量 APP_RUN_MODE ="dev"时，使用 conf/dev.app.conf
//  没有此环境变量时，使用conf/app.conf
func GetAppConfig() *AppConfig {
	//根据环境变量使用不同的conf文件
	runMode := os.Getenv("APP_RUN_MODE")
	if runMode == "" {
		fileName = defaultConfig
	}

	if runMode == "dev" {
		fileName = defaultDevConfig
	}
	if runMode == "prod" {
		fileName = defaultProdConfig
	}

	if fileName == "" {
		fileName = defaultConfig
	}

	config.InitLoad(fileName)

	return &AppConfig{
		AppName:       config.DefaultString("app_name", "gow"),
		RunMode:       config.DefaultString("run_mode", "dev"),
		HttpAddr:      config.DefaultString("http_addr", ":8080"),
		AutoRender:    config.DefaultBool("auto_render", false),
		Views:         config.DefaultString("views", "views"),
		TemplateLeft:  config.DefaultString("template_left", "<<"),
		TemplateRight: config.DefaultString("template_right", ">>"),
	}
}
