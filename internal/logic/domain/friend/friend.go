package friend

import "time"

const (
	FriendStatusApply = 0 // 申请
	FriendStatusAgree = 1 // 同意
)

type Friend struct {
	Id         int64     // 没有用
	UserId     int64     // 当前用户ID (添加好友、同意添加好友)
	FriendId   int64     // 好友ID (添加好友、同意添加好友)
	Remarks    string    // 备注 (添加好友、同意添加好友)
	Extra      string    // 附加字段 (添加好友、同意添加好友)
	Status     int       // 申请状态/同意状态
	CreateTime time.Time // 创建实例时间 (添加好友、同意添加好友)
	UpdateTime time.Time // 更新实例时间 (添加好友、同意添加好友、修改好友信息)
}
