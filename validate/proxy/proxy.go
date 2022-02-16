package proxy

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

// 模拟请求
func GetCurl(hostUrl string, r *http.Request) (response *http.Response, body []byte, err error) {
	// 获取uid
	uidPre, err := r.Cookie("uid")
	if err != nil {
		return
	}

	// 获取sign
	uidSign, err := r.Cookie("sign")
	if err != nil {
		return
	}

	// 模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}

	// 手动指定，排除多余的cookies
	val, _ := url.QueryUnescape(uidSign.Value)
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: val, Path: "/"}
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	// 获取返回结果
	response, err = client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	return
}

// 获取其他节点处理结果
func GetDataFromOtherMap(host string, r *http.Request, port string) bool {

	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, r)
	if err != nil {
		return false
	}

	if response.StatusCode == 200 {
		if string(body) == "200" {
			return true
		} else {
			return false
		}
	}
	return false
}