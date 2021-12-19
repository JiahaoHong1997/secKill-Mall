package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"seckill/common"
	"seckill/datamodels"
	rabbitmq "seckill/rabbitmq"
	"strconv"
	"sync"
)

var (
	// 设置集群地址，最好是内网IP
	hostArray = []string{"192.168.1.11", "192.168.1.11"}
	// 本机IP
	localhost = ""
	// 数量控制接口服务器内网IP
	GetOneIp = "127.0.0.1"
	// 数量控制接口服务器端口
	GetOnePort = "8084"
	// 验证服务器端口
	port = "8083"

	hashConsistent   *common.Consistent
	rabbitMqValidate *rabbitmq.RabbitMQ
	accessControl    = &AccessControl{sourceArray: make(map[int]interface{})}
)

// 用来存放控制信息
type AccessControl struct {
	// 用来存放用户想要存放的信息
	sourceArray map[int]interface{}
	*sync.RWMutex
}

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

// 分布式验证方法，查找响应用户请求的服务器
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
	//uidInt, err := strconv.Atoi(uid)
	//if err != nil {
	//	return false
	//}
	//data := m.GetNewRecord(uidInt)
	//
	//if data != nil {
	//	return true
	//}
	//return false
	return true
}

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
func GetDataFromOtherMap(host string, r *http.Request) bool {

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

func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

func Check(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
	fmt.Println("执行check")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)

	// 获取用户cookie
	userCookie, err := r.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	// 1.分布式权限认证
	right := accessControl.GetDistributedRight(r)
	if right == false {
		w.Write([]byte("false"))
	}

	// 2.获取数量控制权限，防止秒杀出现超卖
	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	// 判断数量控制接口请求状态
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			// 整合下单
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}

			message := datamodels.NewMessage(userID, productID)
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}

			// 生产消息
			err = rabbitMqValidate.PublishSimple(string(byteMessage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

func main() {
	// 采用一致性hash算法设置负载均衡器
	hashConsistent = common.NewConsistent()
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	localIp, err := common.GetIntranceIp()
	if err != nil {
		log.Printf("original err:%v\n", errors.Cause(err))
	}
	fmt.Println(localIp)

	rabbitMqValidate = rabbitmq.NewRabbitMQSimple("secKillProduct")
	defer rabbitMqValidate.Destory()

	// 1.过滤器
	filter := common.NewFilter()
	// 注册拦截器
	filter.RegisterFilterUri("/check", Auth)
	filter.RegisterFilterUri("/checkRight", Auth)
	// 2.启动服务
	http.HandleFunc("/check", filter.Handle(Check))           //
	http.HandleFunc("/checkRight", filter.Handle(CheckRight)) // 验证用户登录权限

	http.ListenAndServe(":8083", nil)
}
