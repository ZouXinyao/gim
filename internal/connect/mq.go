package connect

import (
	"gim/config"
	"gim/pkg/db"
	"gim/pkg/logger"
	"gim/pkg/mq"
	"gim/pkg/pb"
	"time"

	"github.com/go-redis/redis"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// StartSubscribe 启动MQ消息处理逻辑
func StartSubscribe() {
	// 接收Redis订阅的消息，这里的都是房间级的消息
	pushRoomPriorityChannel := db.RedisCli.Subscribe(mq.PushRoomPriorityTopic).Channel()
	pushRoomChannel := db.RedisCli.Subscribe(mq.PushRoomTopic).Channel()
	// 逐个处理收到的消息
	for i := 0; i < config.Connect.PushRoomSubscribeNum; i++ {
		go handlePushRoomMsg(pushRoomPriorityChannel, pushRoomChannel)
	}

	// 这里是全局广播消息
	pushAllChannel := db.RedisCli.Subscribe(mq.PushAllTopic).Channel()
	for i := 0; i < config.Connect.PushAllSubscribeNum; i++ {
		go handlePushAllMsg(pushAllChannel)
	}
}

func handlePushRoomMsg(priorityChannel, channel <-chan *redis.Message) {
	for {
		select {
		case msg := <-priorityChannel:
			handlePushRoom([]byte(msg.Payload))
		default:
			select {
			case msg := <-channel:
				handlePushRoom([]byte(msg.Payload))
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
	}
}

func handlePushAllMsg(channel <-chan *redis.Message) {
	for msg := range channel {
		handlePushAll([]byte(msg.Payload))
	}
}

func handlePushRoom(bytes []byte) {
	var msg pb.PushRoomMsg
	err := proto.Unmarshal(bytes, &msg)
	if err != nil {
		logger.Logger.Error("handlePushRoom error", zap.Error(err))
		return
	}
	// 将消息广播到room中的每个conn
	PushRoom(msg.RoomId, msg.MessageSend)
}

func handlePushAll(bytes []byte) {
	var msg pb.PushAllMsg
	err := proto.Unmarshal(bytes, &msg)
	if err != nil {
		logger.Logger.Error("handlePushRoom error", zap.Error(err))
		return
	}
	// 广播给所有的conn
	PushAll(msg.MessageSend)
}
