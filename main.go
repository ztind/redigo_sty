package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"sync"
	"time"
)
/**
redigo操作redis
 */
func main(){
	//dial()
	//poolDial()
	pubsubConn()
}
func dial(){
	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil {
		panic(err)
	}
	//执行
	execCommmand(conn)

	//关闭链接
	defer conn.Close()
}

func poolDial(){
	//pool封装dail实现连接池
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp","127.0.0.1:6379")
		},
		MaxActive:1,//redis在池中的最大链接数(conn个数)
	}

	//get方法获取conn链接对象
	conn := pool.Get()

	err := conn.Err()
	if err !=nil{
		panic(err)
	}
	//执行
	execCommmand(conn)

	//关闭链接
	defer pool.Close()
}

var wg sync.WaitGroup
//发布订阅
func pubsubConn(){

	const healthCheckPeriod = time.Minute
	c, err := redis.Dial("tcp", "127.0.0.1:6379",
		// Read timeout on server should be greater than ping period.
		//单个命令执行的响应超时时间
		redis.DialReadTimeout(healthCheckPeriod+10*time.Second),
		//单个命令的写入超时时间
		redis.DialWriteTimeout(10*time.Second))

	if err != nil {
		panic(err)
	}

	pubsubConn := redis.PubSubConn{Conn:c}

	//订阅多个频道
	if err := pubsubConn.Subscribe(redis.Args{}.AddFlat("ch1").Add("ch2","ch3")...); err != nil {
		panic(err)
	}

	//链接测试
	err = pubsubConn.Ping("OK")
	if err!=nil {
		panic(err)
	}

	wg.Add(1)
	// Start a goroutine to receive notifications from the server.
	go func() {
		for{
			fmt.Println("--------")
			data := pubsubConn.Receive()//没有消息则阻塞在此
			fmt.Println(data)
		}
		wg.Done()
	}()
	wg.Wait()
	//关闭链接
	defer pubsubConn.Close()
}

func execCommmand(conn redis.Conn){
	//Do方法 发送一条命令到redis服务器端执行并返回结果
	doReply,err := conn.Do("GET","name")
	if err!=nil {
		log.Println(err)
		return
	}

	result,_:=redis.String(doReply,err)
	fmt.Println("Do command : ",result)

	//send发送命令到缓存，flush发送命令到服务端执行,Receive方法获取命令执行结果
	conn.Send("SET","age",18)
	conn.Flush()
	sendReply,err:=conn.Receive()
	send_result,_:=redis.String(sendReply,err)
	fmt.Println("Send-Flush-Receive command : ",send_result)
}