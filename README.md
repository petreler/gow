# gow

gow 是一个基于gin框架思想和beego框架中html模板处理机制，封装的一个go web框架。可用于开发Web API和Web网站项目。


*项目地址：*

```
github.com/gkzy/gow
```

*官网地址：*

```
https://gow.zituo.net （在建中...)
```



### 1. 快速开始

```sh
mkdir hello
cd hello
```

```sh
go mod init
```

```
go get github.com/gkzy/gow
```

##### 1.1 创建 main.go
```go
package main

import (
    "github.com/gkzy/gow"
)

func main() {
    r := gow.Default()

    r.GET("/", func(c *gow.Context) {
        c.JSON(gow.H{
            "code": 0,
            "msg":  "success",
        })
    })

    r.Run()
}

```

##### 1.2 编译和运行
```sh
go build && ./hello
```

##### 1.3 访问地址

```sh
浏览器访问：http://127.0.0.1:8080
```

```sh
curl -i http://127.0.0.1:8080
```

```sh
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Wed, 15 Jul 2020 09:15:31 GMT
Content-Length: 27

{"code":0,"msg":"success"}
```

---

### 2. 做一个网站

包括html模板文件处理和静态资源处理


##### 2.1 目录结构

```
PROJECT_NAME
├──static
      ├── img
            ├──111.jpg
            ├──222.jpg
            ├──333.jpg
      ├──js
      ├──css
├──views
    ├──index.html
    ├──article
        ├──detail.html
├──main.go
```

##### 2.2 设置模板目录

```go
r.SetView("views")
```

##### 2.3 设置静态资源

```go
r.Static("/static", "static")
```

##### 2.4 模板函数


##### 2.5 演示代码

*main.go*

```go
package main

import (
    "github.com/gkzy/gow"
)

func main() {
    r := gow.Default()
    r.AutoRender = true //打开html模板渲染
    r.SetView("views") //默认静态目录为views时，可不调用此方法
    r.StaticFile("favicon.ico","static/img/log.png")  //路由favicon.ico
    r.Static("/static", "static")

    //router
    r.Any("/", IndexHandler)
    r.Any("/article/1", ArticleDetailHandler)


    //自定义hello的模板函数
    //在模板文件可通过 {{hello .string}} 来执行
    r.AddFuncMap("hello", func(str string) string {
        return "hello:" + str
    })

    r.Run()
}

//IndexHandler 首页handler
func IndexHandler(c *gow.Context) {
    c.HTML("index.html", gow.H{
        "name":    "gow",
        "package": "github.com/gkzy/gow",
    })
}

//ArticleDetailHandler 文章详情页handler
func ArticleDetailHandler (c *gow.Context){
    c.HTML("article/detail.html", gow.H{
        "title":    "年薪百万的文科专业有哪些？",
    })
}
```

*views/index.html*

``` sh
<html>
<head>
    <title>{{hello .name}}</title>
    <meta charset="utf-8"/>
</head>
<body>
    <h2>{{.name}}</h2>
    <hr>
    <h5>{{.package}}</h5>
</body>
</html>
```

*views/article/detail.html*

```sh
<html>
<head>
    <title>{{.title}}</title>
    <meta charset="utf-8"/>
    <style>
        img{max-width:600px;}
    </style>
</head>
<body>
    <h2>{{.title}}</h2>
    <hr>
    <p><img src="/static/img/111.jpg"></p>
    <p><img src="/static/img/222.jpg"></p>
    <p><img src="/static/img/333.jpg"></p>
</body>
</html>
```

##### 运行

```sh
go run main.go
或
go build main.go -o app && ./app
```

##### 访问

```sh

https://127.0.0.1:8080/
https://127.0.0.1:8080/article/1

```

---


### 3. 获取值

```go
func GetUser(c *gow.Context){

    //获取字串
    c.GetString("key","default")

    //获取int
    c.GetInt("key",0)

    //获取bool
    c.GetBool("key",false)

    //获取int64
    c.GetInt64("key",-1)

    //获取float
    c.GetFloat("key",0)

    //获取[]string
    var ids []string
    ids = c.GetStrings("ids")  

    //其他方法
    c.GetInt32()
    c.GetInt16()
    c.GetInt8()
    ....
}
```

**获取 request body，并反序列化到 struct**

```go

type User struct {
    Nickname string `json:"nickname"`
    QQ       string `json:"qq"`
}

func GetUser(c *Context){
    user := new(User)
    err := c.DecodeJSONBody(&user)
    if err != nil {
        //handler error
    }

    c.JSON(gow.H{
        "user": user,
    })   

}

```

### 4. 输出值

支持：json xml string html 等方式输出

**JSON**

```go
func GetUser(c *Context){

    //default http.StatusOK
    c.JSON(gow.H{
        "nickname":"gow",
        "age":18,
    })

   //或者，指定 http.StatusCode
    c.ServerJSON(200,gow.H{
        "nickname":"gow",
        "age":18,
    })
}

```

**XML**

```go
func GetUser(c *Context){

    //default http.StatusOK
    c.XML(gow.H{
        "nickname":"gow",
        "age":18,
    })

   //或者，指定 http.StatusCode
    c.ServerXML(200,gow.H{
        "nickname":"gow",
        "age":18,
    })
}

```

**String**

```go
func GetUser(c *Context){

    //default http.StatusOK
    c.String("hello gow...")

   //或者，指定 http.StatusCode
    c.ServerString(200,"hello gow...")
}

```

**File**


```go
func GetUser(c *Context){

    //读取并输出
    c.File("go.mod")
}
```

---
### 5. 路由详解

##### 5.1 支持的HTTPMethod
```go
HTTPMethod = map[string]bool{
    "GET":     true,
    "POST":    true,
    "PUT":     true,
    "DELETE":  true,
    "PATCH":   true,
    "OPTIONS": true,
    "HEAD":    true,
    "TRACE":   true,
}
```

##### 5.2 使用方式

包括基本路由与分组

```go
r := gow.Default()

r.GET(path,handler)
r.POST(path,handler)
r.PUT(path,handler)
r.DELETE(path,handler)
r.PATCH(path,handler)
r.OPTIONS(path,handler)
r.HEAD(path,handler)
r.TRACE(path,handler)
```

##### 5.3 路由参数

```go
r.Any("/article/:id", ArticleDetailHandler)
```

*获取 param 值*

```go
id:=c.Param("id")
```


##### 5.4 路由分组


*main.go*

```go
package main

import (
    "github.com/gkzy/gow"
)

func main() {
    r := gow.Default()
    v1 := r.Group("/v1")
    {
        v1.GET("/user/:id", GetUser)
        v1.DELETE("/user/:id", DeleteUser)
    }

    r.Run()
}

func GetUser(c *gow.Context) {
    c.JSON(gow.H{
        "nickname": "新月却泽滨",
        "qq":       "301109640",
    })
}

func DeleteUser(c *gow.Context) {
    c.JSON(gow.H{
        "code": 0,
        "msg": "success",
    })
}

```

*Get*

```sh
curl -i http://127.0.0.1:8080/v1/user/1


HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 16 Jul 2020 05:55:16 GMT
Content-Length: 46

{
  "nickname": "新月却泽滨",
  "qq": "301109640"
}

```

*Delete*

```sh
curl  -X "DELETE" http://127.0.0.1:8080/v1/user/1

{
    "code":0,
    "msg":"success"
}

```