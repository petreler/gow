# github.com/gkzy/gow/lib/logy

### 基于第三方日志库的封装

```sh
github.com/sirupsen/logrus
```


### 引用

```go
import "github.com/gkzy/gow/lib/logy"

```


### 使用


##### 直接使用


```go
// 输出到控制台

logy.Info("日志内容")
logy.Debug("日志内容")
logy.Error("日志内容")
logy.Trace("日志内容")
logy.Panic("日志内容")
```

##### 存储到文件

```go
//存储到文件

//在调用方初始化一次
logy.InitLogger(&LoggerConfig{
    Path:       "logs",  //目录
    FileName:   "web",  //文件名前缀
    ToFile:     true,   //存储到文件
    MaxDay:     7,      //保存天数
    SplitDay:   1,      //按1天分割
    TimeFormat: "",     //日期格式
})


//然后调用

logy.Info("日志内容")
logy.Debug("日志内容")
logy.Error("日志内容")
logy.Trace("日志内容")
logy.Panic("日志内容")
```