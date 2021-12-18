package main

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"seckill/common"
	"strconv"
	"sync"
)

// 设置集群地址，最好是内网IP
var hostArray = []string{"127.0.0.1", "127.0.0.1"}

// 本机IP
var localhost = "127.0.0.1"

var port = "8080"

var hashConsistent *common.Consistent

// 用来存放控制信息
type AccessControl struct {
	// 用来存放用户想要存放的信息
	sourceArray map[int]interface{}
	*sync.RWMutex
}

var accessControl = &AccessControl{sourceArray: make(map[int]interface{})}

// 获取指定的数据
func (m *AccessControl) GetNewRecord(uid int) interface{} {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()
	data := m.sourceArray[uid]
	return data
}

// 设置记录
func (m *AccessControl) SetNewRecord(uid int) {
	m.RWMutex.Lock()
	m.sourceArray[uid] = "hello world"
	m.RWMutex.Unlock()
}

func (m *AccessControl) GetDistributedRight(r *http.Request) bool {
	// 获取用户uid
	uid, err := r.Cookie("uid")
	if err != nil {
		return false
	}

	// 采用一致性算法，根据用户uid判断获取具体信息的机器
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}

	// 判断是否是本机
	if hostRequest == localhost {
		// 执行本机数据读取和校验
		return m.GetDataFromMap(uid.Value)
	} else {
		// 不是本机则充当代理访问数据返回结果
		return GetDataFromOtherMap(hostRequest, r)
	}
}

// 获取本机map，并且处理业务逻辑，返回结果是bool类型
func (m *AccessControl) GetDataFromMap(uid string) (isOk bool) {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}
	data := m.GetNewRecord(uidInt)

	if data != nil {
		return true
	}
	return false
}

// 获取其他节点处理结果
func GetDataFromOtherMap(host string, r *http.Request) bool {
	// 获取uid
	uidPre, err := r.Cookie("uid")
	if err != nil {
		return false
	}

	// 获取sign
	uidSign, err := r.Cookie("sign")
	if err != nil {
		return false
	}

	// 模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+host+":"+port+"/access", nil)
	if err != nil {
		return false
	}

	// 手动指定，排除多余的cookies
	val, _ := url.QueryUnescape(uidSign.Value)
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: val, Path: "/"}
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	// 获取返回结果
	response, err := client.Do(req)
	if err != nil {
		return false
	}
	body, err := ioutil.ReadAll(response.Body)
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

// 统一验证拦截器，每个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	err := CheckUserInfo(r)
	if err != nil {
		return err
	}
	return nil
}

// 身份校验函数
func CheckUserInfo(r *http.Request) error {
	// 1.获取uid cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("uid got failed")
	}

	// 2.获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("sign got failed")
	}

	val, _ := url.QueryUnescape(signCookie.Value)
	signByte, err := common.DePwdCode(val)
	if err != nil {
		return errors.New("加密串已被串改")
	}

	if checkInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败！")
}

func checkInfo(checkStr string, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}

func Check(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
	fmt.Println("执行check")
}

func main() {
	// 采用一致性hash算法设置负载均衡器
	hashConsistent = common.NewConsistent()
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	// 1.过滤器
	filter := common.NewFilter()
	// 注册拦截器
	filter.RegisterFilterUri("/check", Auth)
	// 2.启动服务
	http.HandleFunc("/check", filter.Handle(Check))

	http.ListenAndServe(":8083", nil)
}
