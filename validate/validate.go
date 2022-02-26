package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"net/url"
	"seckill/common"
	"seckill/common/loadBalance"
	"seckill/common/lock"
	"seckill/common/rabbitmq"
	"seckill/dao/db"
	"seckill/models"
	"seckill/validate/auth"
	"seckill/validate/proxy"
	"seckill/validate/tokenLimit"
	"strconv"
	"sync"
	"time"
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

	hashConsistent   loadBalance.LoadBalance
	rabbitMqValidate *RabbitMQ.RabbitMQ
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
		return proxy.GetDataFromOtherMap(hostRequest, r, port)
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

func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		w.WriteHeader(502)
		return
	}
	w.Write([]byte("true"))
	return
}

func publishMessage(userID, productID int64) error {
	message := models.NewMessage(userID, productID)
	byteMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = rabbitMqValidate.PublishSimple(string(byteMessage))
	if err != nil {
		return err
	}
	return nil
}

func CheckLocal(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
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
		return
	}

	// 2.获取数量控制权限，防止秒杀出现超卖
	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := proxy.GetCurl(hostUrl, r)
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

			err = publishMessage(userID, productID)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 生产消息
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

func CheckCache(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

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
		return
	}

	lockPool := lock.NewRedisPool()
	proPool := db.NewCachePool()

	if !lockPool.Lock(ctx) {
		w.Write([]byte("overtime"))
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}
	defer lockPool.UnLock()
	num, err := proPool.HGet(lockPool.Ctx, productString, "productInventory").Result()
	if err != nil {
		w.Write([]byte("cannot get product nums"))
		return
	}
	proNum, _ := strconv.ParseInt(num,10,64)
	if proNum > 0 {
		_, err := proPool.HIncrBy(lockPool.Ctx, productString, "productInventory",-1).Result()
		if err != nil {
			w.Write([]byte("decr product failed"))
			return
		}
		productID, _ := strconv.ParseInt(productString, 10, 64)
		userID, _ := strconv.ParseInt(userCookie.Value, 10, 64)
		err = publishMessage(userID, productID)
		if err != nil {
			_, err = proPool.HIncrBy(lockPool.Ctx, productString, "productInventory", 1).Result()
			w.Write([]byte("generate order failed"))
			return
		}
	} else {
		w.Write([]byte("no more product"))
		return
	}
}

func main() {
	// 采用一致性hash算法设置负载均衡器
	hashConsistent = loadBalance.LoadBalanceFactory(loadBalance.LbConsistent, 20)
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	localIp, err := common.GetIntranceIp()
	if err != nil {
		log.Printf("original err:%v\n", errors.Cause(err))
	}
	fmt.Println(localIp)

	// 初始化消息队列
	rabbitMqValidate = RabbitMQ.NewRabbitMQSimple("secKillProduct")
	defer rabbitMqValidate.Destory()

	// 1.初始化拦截器
	filter := common.NewFilter()

	// 注册拦截器验证用户是否登录
	filter.RegisterFilterUri("/checkLocal", auth.Auth)
	filter.RegisterFilterUri("/checkCache", auth.Auth)
	filter.RegisterFilterUri("/checkRight", auth.Auth)

	// 注册拦截器验证请求是否通过了令牌桶算法的拦截
	filter.RegisterFilterUri("/checkLocal", tokenLimit.LimitT)
	filter.RegisterFilterUri("/checkCache", tokenLimit.LimitT)

	// 2.启动服务
	http.HandleFunc("/checkLocal", filter.Handle(CheckLocal))           // 超热点商品在本地内存进行数量控制
	http.HandleFunc("/checkCache", filter.Handle(CheckCache))		   // 直接在缓存层进行数量控制，成功下单后再修改数据库
	http.HandleFunc("/checkRight", filter.Handle(CheckRight)) // 验证用户登录权限

	http.ListenAndServe(":8083", nil)
}
