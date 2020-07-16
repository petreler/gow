package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gkzy/gow/lib/util"
	"io"
	"net/http"
	"time"
)

//AliClient AliClient
type AliClient struct {
	AccessKeyId string
	Secret      string
	EndPoint    string
	BucketName  string
	ServerUrl   string
}

//NewAliClient NewAliClient
func (c *AliClient) NewAliClient(accessKeyId, secret, endPoint, bucketName, serverUrl string) *AliClient {
	//末尾添加/
	if len(serverUrl) > 0 && serverUrl[len(serverUrl)-1:] != "/" {
		serverUrl = serverUrl + "/"
	}
	return &AliClient{
		AccessKeyId: accessKeyId,
		Secret:      secret,
		EndPoint:    endPoint,
		BucketName:  bucketName,
		ServerUrl:   serverUrl,
	}
}

//UploadPhoto 上传一个图片文件到aliyun oss
func (c *AliClient) UploadPhoto(reader io.Reader, dir, suffix string) (url string, err error) {
	if suffix == "" {
		suffix = ".jpg"
	}
	client, err := oss.New(c.EndPoint, c.AccessKeyId, c.Secret)
	if err != nil {
		return
	}
	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return
	}
	uuid, _ := util.GetUUID()
	imgUrl := fmt.Sprintf("%s/%s/%s", dir, time.Now().Format("20060102"), uuid+suffix)
	err = bucket.PutObject(imgUrl, reader)
	if err != nil {
		return
	}
	url = fmt.Sprintf("%s%s", c.ServerUrl, imgUrl)
	return
}

//UploadHttpPhoto 上传网络/远程图片到oss
func (c *AliClient) UploadHttpPhoto(rUrl, dir string) (url string, err error) {
	resp, err := http.Get(rUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return c.UploadPhoto(resp.Body, dir, "")
}
