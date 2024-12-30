package main

import (
	"fmt"
	"middleproject/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	// POST 请求：发验证码
	r.POST("/api/regist/mail", func(c *gin.Context) {
		fmt.Println("收到发验证码请求")
		service.SendMailInterface(c)
	})
	r.POST("/api/regist", func(c *gin.Context) {
		fmt.Println("收到注册请求")
		service.Register(c)
	})
	//发帖
	r.POST("/api/newlog", func(c *gin.Context) {
		fmt.Println("收到登录请求")
		service.PublishPost(c)
	})
	//发评论
	r.POST("/api/newcomment", func(c *gin.Context) {
		fmt.Println("收到评论请求")
		service.PublishComment(c)
	})
	//发回复
	r.POST("/api/newreply", func(c *gin.Context) {
		fmt.Println("收到回复请求")
		service.PublishReply(c)
	})
	r.GET("/api/logs", func(c *gin.Context) {
		fmt.Println("收到获取帖子请求")
		service.GetRecommendPost(c)
	})
	//获取帖子详情
	r.GET("/api/logs/alog", func(c *gin.Context) {
		fmt.Println("收到获取帖子详情请求")
		service.GetPostInfo(c)
	})
	//获取帖子图片
	r.GET("/api/logs/image", func(c *gin.Context) {
		fmt.Println("收到获取帖子图片请求")
		service.GetPostImage(c)
	})
	//获取更多评论
	r.GET("/api/morecom", func(c *gin.Context) {
		fmt.Println("收到获取更多评论请求")
		service.GetMoreComment(c)
	})
	//获取更多回复
	r.GET("/api/morereply", func(c *gin.Context) {
		fmt.Println("收到获取更多回复请求")
		service.GetMoreReply(c)
	})
	//搜索帖子
	r.POST("/api/searchlogs", func(c *gin.Context) {
		fmt.Println("收到搜索帖子请求")
		service.SearchPost(c)
	})
	//管理员登录
	r.POST("/api/admlogin", func(c *gin.Context) {
		fmt.Println("收到管理员登录请求")
		service.AdmLogin(c)
	})
	//获取所有用户信息
	r.GET("/api/adm/alluser", func(c *gin.Context) {
		fmt.Println("收到获取用户信息请求")
		service.GetallUser(c)
	})
	//获取所有帖子信息
	r.GET("/api/adm/allpost", func(c *gin.Context) {
		fmt.Println("收到获取帖子信息请求")
		service.GetallPost(c)
	})
	//管理员搜索用户
	r.GET("/api/adm/searchUser", func(c *gin.Context) {
		fmt.Println("收到搜索用户请求")
		service.AdmSearchUser(c)
	})
	//管理员搜索帖子
	r.GET("/api/adm/searchLog", func(c *gin.Context) {
		fmt.Println("收到搜索帖子请求")
		service.AdmSearchPost(c)
	})
	//管理员删除贴子
	r.POST("/api/adm/admdellog", func(c *gin.Context) {
		fmt.Println("收到删除帖子请求")
		service.AdmDeletePost(c)
	})
	//管理员删除评论
	r.POST("/api/adm/admdelcomment", func(c *gin.Context) {
		fmt.Println("收到删除评论请求")
		service.AdmDeleteComment(c)
	})
	//管理员删除回复
	r.POST("/api/adm/admdelreply", func(c *gin.Context) {
		fmt.Println("收到删除回复请求")
		service.AdmDeleteReply(c)
	})
	//管理员封禁禁言用户
	r.POST("/api/adm/gagandenclose", func(c *gin.Context) {
		fmt.Println("收到封禁禁言用户请求")
		service.AdmBan(c)
	})
	//管理员警告用户
	r.POST("/api/adm/sendinfo", func(c *gin.Context) {
		fmt.Println("收到警告用户请求")
		service.AdmWarn(c)
	})
	//管理员忽略举报
	r.POST("/api/adm/reportok", func(c *gin.Context) {
		fmt.Println("收到忽略举报请求")
		service.AdmIgnore(c)
	})
	//用户获取系统信息
	r.GET("/api/adm/getmessage", func(c *gin.Context) {
		fmt.Println("收到获取系统信息请求")
		service.Getsysinfo(c)
	})
	//登录
	r.POST("/api/login", func(c *gin.Context) {
		fmt.Println("收到登录请求")
		service.Login(c)
	})
	//管理员仅仅警告
	r.POST("/api/adm/dWarn", func(c *gin.Context) {
		fmt.Println("收到警告请求")
		service.AdmOnlyWarn(c)
	})
	//管理员仅禁言/封禁
	r.POST("/api/adm/dBan", func(c *gin.Context) {
		fmt.Println("收到禁言/封禁请求")
		service.AdmOnlyBan(c)
	})

	//获取个人设置接口
	r.GET("/api/persetting", service.HandleGetPersonalSettings)

	//更新个人设置接口
	r.POST("/api/persetting/edit", service.UpdatePersonalSettings)

	//忘记密码
	r.POST("/api/forget", service.ForgotPassword)

	//获取个人信息
	r.GET("/api/perinfo", service.HandleGetPersonalInfo)
	//修改个人信息
	r.POST("/api/perinfo/edit", service.UpdatePersonalInfo)
	//点赞帖子
	r.POST("/api/likelog", service.LikePost)
	//收藏帖子
	r.POST("/api/collectlog", service.CollectPost)
	//喜欢评论
	r.POST("/api/likecomment", service.LikeComment)
	//关注用户
	r.POST("/api/attionusr", service.HandleFollowAction)
	//喜欢回复
	r.POST("/api/likereply", service.LikeReply)
	//举报接口
	r.POST("/api/report", service.HandleReport)
	//获取个人发帖
	r.GET("/api/perlogs", service.GetPersonalPostLogs)

	//获取个人喜欢帖子
	r.GET("/api/perlikelogs", service.GetPersonalLikePosts)

	//获取个人收藏帖子
	r.GET("/api/percollectlogs", service.GetPersonalCollectPosts)

	//删帖
	r.POST("/api/deletelog", service.DeletePost)
	//删除回复
	r.POST("/api/deletereply", service.DeleteReply)
	//删除评论
	r.POST("/api/deletecomment", service.DeleteComment)

	//获取举报目标
	r.GET("/api/adm/getreports", service.GetReports)
	//获取举报目标详情
	r.GET("/api/adm/getreportinfo", service.GetReportInfo)

	// 获取某个用户的所有粉丝
	r.GET("/api/per/attioned", service.GetFollowers)
	// 获取某个用户关注了哪些其他用户
	r.GET("/api/per/attion", service.GetFollowing)
	//解除封禁/禁言
	r.POST("/api/adm/liftBan", service.HandleUnmute)
	//修改被限制(封禁，禁言)时间
	r.POST("/api/adm/modLiftDays", service.HandleUpdateMuteTime)

	//获取用户状态
	r.GET("/api/adm/getUserStatus", service.GetUserStatus)
	// 启动 HTTP 服务器
	if err := r.Run(":8080"); err != nil {
		fmt.Println("启动服务器失败:", err)
	}
}
