package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	driverMysql "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
	Red *redis.Client
)

func InitApp() {
	// fmt.Println("init app")
	// 设置加载的配置文件的名称
	viper.SetConfigName("app")
	// 设置加载的配置文件的位置
	viper.AddConfigPath("config")
	// 加载配置文件的内容
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("config mysql:", viper.Get("mysql"))
}

func InitMysql() error {
	// 自定义日志模板，打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //慢SQL阈值
			LogLevel:      logger.Info, //级别
			Colorful:      true,        //彩色
		},
	)

	port,_ := strconv.Atoi(viper.GetString("mysql.port"))

	dbConfig := driverMysql.Config{
		User: viper.GetString("mysql.user"),
		Passwd: viper.GetString("mysql.password"),
		Net: "tcp",
		Addr: fmt.Sprintf("%s:%d",viper.GetString("mysql.ip"),port),
		DBName: viper.GetString("mysql.database"),
		ParseTime: true,
		Loc:time.Local,
		Params: map[string]string{
			"charset":"utf8mb4",
		},
	}

	dsn := dbConfig.FormatDSN()

	// fmt.Println("init mysql")
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return err
	}

	// user := models.UserBasic{}
	// res := DB.Find(&user)

	// fmt.Println(res)
	return nil
}

func InitRedis(){
	Red = redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.Addr"),
		DB: viper.GetInt("redis.DB"),
		Password: viper.GetString("redis.password"),
		PoolSize: viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleCoon"),
	})

	pong,err := Red.Ping(context.Background()).Result()

	if err != nil{
		fmt.Println("【redis】init redis error ....",err)
	}else{
		fmt.Println("【redis】ping redis is success ...",pong)
	}
}

var PublishKey = "publish_key"

// 发布消息到redis上
func Publish(ctx context.Context,channel string,msg string) error {
	var err error
	fmt.Println("publish message ...",msg)
	err = Red.Publish(ctx,channel,msg).Err()
	return err
}

// 订阅redis上的消息
func Subscribe(ctx context.Context,channel string) (string, error){
	sub := Red.Subscribe(ctx,channel)
	msg,err := sub.ReceiveMessage(ctx)
	fmt.Println("Subscribe...",msg.Payload)
	return msg.Payload,err
}
