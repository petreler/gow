package gow

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
