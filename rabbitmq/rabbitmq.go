package RabbitMQ

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"seckill/datamodels"
	"seckill/service"
	"sync"

	"github.com/streadway/amqp"
)

// url格式: amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://JiahaoHong1997:hjh1314lvwxy@127.0.0.1:5672/hjh"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string // 队列名称
	Exchange  string // 交换机
	Key       string // key
	Mqurl     string // 连接信息
	sync.Mutex
}

// 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}

	var err error
	// 创建rabbitmq连接
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "Connect Failed!")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "Fetch channel Failed!")
	return rabbitmq
}

// 断开channel和connection
func (r *RabbitMQ) Destory() {
	r.channel.Close()
	r.conn.Close()
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
	}
}

// 只使用默认交换机情况下的实例创建、发送和接收方法
// step1:创建Simple模式下的RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "") // Simple模式下只使用queque，是有默认交换机的（direct）
}

// step2:Simple模式下生产者代码
func (r *RabbitMQ) PublishSimple(message string) error {
	r.Lock()
	defer r.Unlock()
	// 1.申请队列，如果队列不存在会自动创建，如果存在则跳过创建。保证队列存在消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性
		false, // 是否阻塞
		nil,   // 额外属性
	)
	if err != nil {
		return errors.Wrap(err, "QueueDeclare failed")
	}

	// 2.发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		false, // 如果为true，根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会吧发送的消息回退给发送者
		false, // 如果为true，当exchange发送消息到队列后发现队列上没有绑定消费者，则会吧消息发还给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	return nil
}

// step3:Simple模式下消费者代码
func (r *RabbitMQ) ConsumeSimple(orderService service.IOrderService, productService service.IProductService) {
	// 1.申请队列，如果队列不存在会自动创建，如果存在则跳过创建。保证队列存在消息能发送到队列中
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性
		false, // 是否阻塞
		nil,   // 额外属性
	)
	if err != nil {
		fmt.Println(err)
	}

	// 消费者流量控制
	r.channel.Qos(
		1,     // 当前消费者一次能接收最大消息数量
		0,     // 服务器传递的最大容量（以8位字节为单位）
		false) // 如果设置为true，对channel可用

	// 2.接受消息
	msgs, err := r.channel.Consume(
		q.Name,
		"",    // 用来区分多个消费者
		false, // 是否自动应答
		false, // 是否具有排他性
		false, // 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false, // 队列消费是否阻塞（false为阻塞）
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	// 启用协程处理消息
	go func() {
		for d := range msgs {
			// 实现我们要处理的逻辑函数
			log.Printf("Recieve a message: %s", d.Body)
			message := &datamodels.Message{}
			err := json.Unmarshal([]byte(d.Body), message)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(message)

			// 插入订单
			_, err = orderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}
			// 扣除商品数量
			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}

			d.Ack(false) // 为false，告诉rabbitMq当前消息已经被消费完，如果为true，表示确实所有未确认的消息
		}
	}()

	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	<-forever
}
