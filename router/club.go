package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qipai/dao"
	"qipai/enum"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"strconv"
)

func club() {
	r := R.Group("/clubs")
	r.Use(middleware.JWTAuth())
	// 创建俱乐部
	r.POST("", clubCreateFunc)
	// 获取指定俱乐部信息
	r.GET("/:cid", clubGetFunc)
	// 列出加入的俱乐部
	r.GET("", clubsFunc)
	// 解散俱乐部
	r.DELETE("/:cid", clubDeleteFunc)
	// 修改俱乐部名称和公告
	r.PUT("/:cid", clubEditFunc)
	// 加入俱乐部
	r.POST("/:cid/user", clubJoinFunc)
	// /1/users会员列表  /1/users?verify 待审核会员列表
	r.GET("/:cid/users/*type", clubUsersFunc)
	// 编辑会员状态：设为管理 取消管理  冻结 取消冻结 设为代付 取消代付 审核通过用户  移除用户
	r.PUT("/:cid/user/:uid/*action", clubEditUserFunc) // 这里做了路由修改，对应的功能代码也需要修改，记录一下，修改好后，删除该注释
}

func clubCreateFunc(c *gin.Context) {
	type ClubForm struct {
		Score     enum.ScoreType `form:"score" json:"score"`                    // 底分方式
		Players   int            `form:"players" json:"players"`                // 玩家个数
		Count     int            `form:"count" json:"count" binding:"required"` // 局数
		StartType enum.StartType `form:"start" json:"start"`                    // 0 第一个入场的开始  1 全准备好开始
		Pay       enum.PayType   `form:"pay" json:"pay"`                        // 0 房主  1 AA
		King      enum.KingType  `form:"king" json:"king"`                      // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
		Special   int            `form:"special" json:"special"`                // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
		Times     enum.TimesType `form:"times" json:"times"`                    // 翻倍规则
	}

	var form ClubForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	// 限制只能 10  20 30 局
	if form.Count != 10 && form.Count != 20 && form.Count != 30 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-2).Msg("局数[count]只能是10/20/30"))
		return
	}

	// 限制游戏开始方式
	if form.StartType != 0 && form.StartType != 1 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-3).Msg("开始方式[start]只能是0或1"))
		return
	}

	// 限制支付模式
	if form.Pay != 0 && form.Pay != 1 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-4).Msg("支付方式[pay]只能是0或1"))
		return
	}
	// 限制王癞模式
	if form.King != 0 && form.King != 1 && form.King != 2 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-5).Msg("王癞模式[king]只能是0/1/2"))
		return
	}

	// 限制特殊牌型 全部选中状态为7位2进制都是1，最大为1111111==127
	if form.Special > 127 || form.Special < 0 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-6).Msg("特殊牌型取值不合法"))
		return
	}

	// 限制翻倍规则
	if form.Times < 0 || form.Times > 4 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-7).Msg("翻倍规则[times]取值不合法，只能在0-4之间"))
		return
	}

	// 底分取值不合法
	if form.Score < 0 || form.Score > 5 {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-7).Msg("底分类型取值只能在0-5之间"))
		return
	}

	info := c.MustGet("user").(*utils.UserInfo)
	club := &model.Club{}
	club.Special = form.Special
	club.King = form.King
	club.Pay = form.Pay
	club.StartType = form.StartType
	club.Count = form.Count
	club.Score = form.Score
	club.Times = form.Times

	// 如果是老板支付，就默认需要审核才能进入俱乐部
	if club.Pay == enum.PayBoss {
		club.Check = true
	} else if club.Pay == enum.PayAA {
		club.Check = false
	}

	club.Uid = info.Uid

	if err := srv.Club.CreateClub(club); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-7).Msg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Msg().Msg("创建成功").AddData("id", club.ID))
}

func clubGetFunc(c *gin.Context) {

	type clubV struct {
		Id        uint           `json:"id"`
		Name      string         `json:"name"`       // 俱乐部名称
		Check     bool           `json:"check"`      // 是否审查
		Notice    string         `json:"notice"`     // 公告
		RollText  string         `json:"roll_text"`  // 俱乐部大厅滚动文字
		Score     enum.ScoreType `json:"score"`      // 底分 以竖线分割的底分方式
		Players   int            `json:"players"`    // 玩家个数
		Count     int            `json:"count"`      // 局数
		StartType enum.StartType `json:"start_type"` // 游戏开始方式
		Pay       enum.PayType   `json:"pay"`        // 付款方式 0 俱乐部老板付 1 AA
		Times     enum.TimesType `json:"times"`      // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
		Special   int            `json:"special"`    // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
		King      enum.KingType  `json:"king"`       // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
		Uid       uint           `json:"uid"`        // 老板
		Close     bool           `json:"close"`      // 是否打烊
	}

	info := c.MustGet("user").(*utils.UserInfo)
	idStr, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("参数格式不合法"))
		return
	}

	club, err := srv.Club.GetClub(info.Uid, uint(idStr))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	var cv clubV

	if !utils.Copy(club, &cv) {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("内容转换出错"))
		return
	}

	c.JSON(http.StatusOK, utils.Msg().AddData("club", cv))
}

func clubsFunc(c *gin.Context) {
	type clubV struct {
		Id      uint           `json:"id"`
		Score   enum.ScoreType `json:"score"`
		Pay     enum.PayType   `json:"pay"`
		Count   int            `json:"count"`
		Boss    string         `json:"boss"`
		BossUid uint           `json:"bossUid"`
	}

	info := c.MustGet("user").(*utils.UserInfo)

	var clubsV []clubV
	for _, v := range srv.Club.MyClubs(info.Uid) {
		var u model.User
		dao.Db.First(&u, v.Uid)
		clubsV = append(clubsV, clubV{
			Id:      v.ID,
			Score:   v.Score,
			Pay:     v.Pay,
			Count:   v.Count,
			Boss:    u.Nick,
			BossUid: u.ID,
		})
	}
	c.JSON(http.StatusOK, utils.Msg().AddData("clubs", clubsV))
}

func clubDeleteFunc(c *gin.Context) {
	club := &model.Club{}
	id, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("俱乐部编号不正确"))
		return
	}
	club.ID = uint(id)
	dao.Db.Delete(club)
	c.JSON(http.StatusOK, utils.Msg().Msg("解散成功"))
}

func clubEditFunc(c *gin.Context) {
	type infoFrom struct {
		Check  bool   `form:"check" json:"check"`
		Close  bool   `form:"close" json:"close"`
		Name   string `form:"name" json:"name"`
		Notice string `form:"notice" json:"notice"`
	}
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("俱乐部编号不正确"))
		return
	}

	var form infoFrom
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	if err := srv.Club.UpdateNameAndNotice(uint(cid), form.Check,form.Close,form.Name, form.Notice); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Msg().Msg("编辑成功"))
}

func clubJoinFunc(c *gin.Context) {
	info := c.MustGet("user").(*utils.UserInfo)
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("加入失败，俱乐部编号只能是数字"))
		return
	}

	err = srv.Club.Join(uint(cid), info.Uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Msg().Msg("加入成功"))
}

func clubUsersFunc(c *gin.Context) {

	info := c.MustGet("user").(*utils.UserInfo)
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("俱乐部编号只能是数字"))
		return
	}
	// 只能看到自己加入的俱乐部的用户列表
	if !srv.Club.IsClubUser(info.Uid, uint(cid)) {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("你不属于该俱乐部，无法查看该俱乐部用户列表"))
		return
	}

	users := srv.Club.Users(uint(cid))

	c.JSON(http.StatusOK, utils.Msg().Msg("获取俱乐部用户列表成功").AddData("users", users))
}

func clubEditUserFunc(c *gin.Context) {
	type formAction struct {
		// 编辑会员状态：设为管理(admin) 取消管理(-admin)  冻结(disable) 取消冻结(-disable) 设为代付(pay) 取消代付(-pay) 审核通过用户(check)  移除用户(remove)
		Action string `form:"action" json:"action" binding:"required"`
		Uid    uint   `form:"uid" json:"uid" binding:"required"`
		Cid    uint   `form:"cid" json:"cid" binding:"required"`
	}
	var form formAction
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)

	isAdmin := srv.Club.IsAdmin(info.Uid, form.Cid)
	isBoss := srv.Club.IsBoss(info.Uid, form.Cid)
	// 只有管理员或创建者可以操作
	if !isAdmin && !isBoss {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是管理员或老板，无法操作！"))
		return
	}

	// 自己不能编辑自己
	if info.Uid == form.Uid {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不能对自己进行操作！"))
		return
	}

	switch form.Action {
	case "admin":
		// 只有老板可以设置管理员
		if !isBoss {
			c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是老板，无法设置管理员！"))
			return
		}
		err = srv.Club.SetAdmin(form.Cid, form.Uid, true)
	case "-admin":
		// 只有老板可以取消管理员
		if !isBoss {
			c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是老板，无法取消管理员！"))
			return
		}
		err = srv.Club.SetAdmin(form.Cid, form.Uid, false)
	case "disable":
		err = srv.Club.SetDisable(form.Cid, form.Uid, true)
	case "-disable":
		err = srv.Club.SetDisable(form.Cid, form.Uid, false)
	case "pay":
		if !isBoss {
			c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是老板，无法设置代付！"))
			return
		}
		err = srv.Club.SetPay(form.Cid, form.Uid, true)
	case "-pay":
		if !isBoss {
			c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是老板，无法取消代付！"))
			return
		}
		err = srv.Club.SetPay(form.Cid, form.Uid, false)
	case "check":
		// 审核通过，就是设置为普通用户，跟取消冻结操作一样
		err = srv.Club.SetDisable(form.Cid, form.Uid, false)
	case "remove":
		if !isBoss {
			c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("您不是老板，无法取消代付！"))
			return
		}
		srv.Club.RemoveClubUser(form.Cid, form.Uid)
	default:
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("不支持这个操作:"+form.Action))
	}
	c.JSON(http.StatusOK, utils.Msg().Msg("操作成功"))
}
