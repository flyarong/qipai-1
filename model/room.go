package model

import (
	"github.com/jinzhu/gorm"
	"qipai/enum"
	"time"
)

type Club struct {
	Name     string // 俱乐部名称
	Check    bool   // 是否审查
	Notice   string // 公告
	RollText string // 俱乐部大厅滚动文字
	Close    bool   // 是否打烊
	PayerUid uint   // 代付用户id
	ClubRoomBase
}

// 提取club和room共有的字段
type ClubRoomBase struct {
	gorm.Model
	Score     enum.ScoreType // 底分 以竖线分割的底分方式
	Players   int            // 玩家个数
	Count     int            // 总局数
	StartType enum.StartType // 游戏开始方式
	Pay       enum.PayType   // 付款方式 0 俱乐部老板付 1 AA
	Times     enum.TimesType // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
	Special   int            // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
	King      enum.KingType  // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
	Uid       uint           // 老板
	BossNick  string         // 老板昵称
	Tui       bool           // 是否推注
}

type Room struct {
	ClubRoomBase
	Current int             // 当前第几局
	Status  enum.GameStatus // 0 未开始，1 游戏中， 2 已结束
	ClubId  uint            // 属于哪个俱乐部
	TableId int             // 俱乐部第几桌
}

// 记录茶楼每一桌的属性，创建茶楼房间的时候，优先根据这个属性创建房间
type ClubRoom struct {
	ClubRoomBase
	ClubId  uint // 属于哪个俱乐部
	TableId int  // 俱乐部第几桌
}

// 记录俱乐部的用户
type ClubUser struct {
	gorm.Model
	Uid    uint              // 用户编号
	ClubId uint              // 俱乐部编号
	Status enum.ClubUserType // 0 等待审核，1 正式用户， 2 冻结用户
	Admin  bool              // 是否是管理员 true 是管理员
}

// 记录房间中的用户
type Player struct {
	gorm.Model
	Uid        uint                     // 用户编号
	Nick       string                   // 昵称
	Avatar     string                   // 用户头像
	DeskId     int                      // 座位号
	RoomId     uint                     // 房间编号
	TotalScore int                      // 当前房间中游戏总分
	JoinedAt   *time.Time `sql:"index"` // 加入时间
}

type Game struct {
	gorm.Model
	Banker     bool   `json:"banker"`                  // 是否是庄家 true表示是庄家
	PlayerId   uint   `json:"playerId"`                // 玩家编号
	RoomId     uint   `json:"roomId"`                  // 房间编号
	DeskId     int    `json:"deskId"`                  // 座位号
	Score      int    `json:"score"`                   // 闲家下注
	Times      int    `gorm:"default:-1" json:"times"` // 抢庄倍数
	CardType   int    `json:"cardType"`                // 牌类型，记录是牛几
	Special    int    `json:"special"`                 // 特殊牌型加倍
	TotalScore int    `json:"totalScore"`              // 输赢积分，通过底分*庄家倍数*特殊牌型加倍 计算
	Cards      string `json:"cards"`                   // 用户所拥有的牌
	Current    int    `json:"current"`                 // 这是第几局
	Auto       bool   `json:"auto"`                    // 是否自动托管
	Tui        bool   `json:"tui"`                     // 当局是否推注
}
