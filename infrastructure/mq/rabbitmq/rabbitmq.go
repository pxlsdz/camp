package rabbitmq

import (
	"camp/infrastructure/stores/mysql"
	"camp/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"log"
	"sync"
)

var rabbitmq *RabbitMQ

func GetRabbitMQ() *RabbitMQ {
	return rabbitmq
}

// RabbitMQ rabbitMQ结构体
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机名称
	Exchange string
	//bind Key 名称
	Key string
	//连接信息
	Mqurl string
	sync.Mutex
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	addr := viper.GetString("rabbitMQ.addr")
	username := viper.GetString("rabbitMQ.username")
	password := viper.GetString("rabbitMQ.password")
	// MQURL 连接信息
	MQUrl := fmt.Sprintf("amqp://%s:%s@%s/", username, password, addr)

	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: MQUrl}
}

// Destory 断开channel 和 connection
func (r *RabbitMQ) Destory() {
	r.channel.Close()
	r.conn.Close()
}

//错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	var err error
	//创建RabbitMQ实例
	rabbitmq = NewRabbitMQ(queueName, "", "")
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// PublishSimple 直接模式队列生产
func (r *RabbitMQ) PublishSimple(message string) error {
	r.Lock()
	defer r.Unlock()
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		return err
	}
	//调用channel 发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		//如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	return nil
}

// ConsumeSimple simple 模式下消费者
func (r *RabbitMQ) ConsumeSimple() {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	//消费者流控
	r.channel.Qos(
		1,     //当前消费者一次能接受的最大消息数量
		0,     //服务器传递的最大容量（以八位字节为单位）
		false, //如果设置为true 对channel可用
	)

	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, // queue
		//用来区分多个消费者
		"", // consumer
		//是否自动应答
		//这里要改掉，抢课用手动应答
		false, // auto-ack
		//是否独有
		false, // exclusive
		//设置为true，表示 不能将同一个Conenction中生产者发送的消息传递给这个Connection中 的消费者
		false, // no-local
		//列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		fmt.Println(err)
	}

	//forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d := range msgs {

			//log.Printf("Received a message: %s", d.Body)
			studentCourse := &models.StudentCourse{}
			json.Unmarshal([]byte(d.Body), studentCourse)

			db := mysql.GetDb()
			err = db.Transaction(func(tx *gorm.DB) error {
				// 扣除课程数量
				course := models.Course{ID: studentCourse.CourseID}
				result := tx.Model(&course).Where("cap > 0").UpdateColumn("cap", gorm.Expr("cap - ?", 1))
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return errors.New("无该课程或者库存不足")
				}

				// 创建课程记录
				if err := tx.Create(&studentCourse).Error; err != nil {
					return err
				}
				// 返回 nil 提交事务
				return nil
			})
			//如果为true表示确认所有未确认的消息，
			//为false表示确认当前消息
			d.Ack(false)
		}
	}()

	//log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	//<-forever

}
