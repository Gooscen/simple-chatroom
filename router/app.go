/**
* @Auth:ShenZ
* @Description:
* @CreateDate:2022/06/14 11:57:55
 */
package router

import (
	"simple-chatroom/docs"
	"simple-chatroom/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	// 初始化 JWT
	service.InitJWT()

	r := gin.Default()
	//swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//静态资源
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
	//	r.StaticFS()
	r.LoadHTMLGlob("views/**/*")

	// 公开路由
	public := r.Group("/")
	{
		//首页
		public.GET("/", service.GetIndex)
		public.GET("/index", service.GetIndex)
		public.GET("/toRegister", service.ToRegister)
		public.GET("/toChat", service.ToChat)
		public.GET("/chat", service.Chat)

		// 认证相关
		public.POST("/login", service.LoginHandler)
		public.POST("/logout", service.LogoutHandler)
		public.POST("/refresh_token", service.RefreshHandler)

		public.POST("/user/createUser", service.CreateUser)
		public.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	}

	// 需要认证的路由
	auth := r.Group("/")
	auth.Use(service.JWTAuth())
	{
		//用户模块
		auth.POST("/user/getUserList", service.GetUserList)
		//auth.POST("/user/createUser", service.CreateUser)
		auth.POST("/user/deleteUser", service.DeleteUser)
		auth.POST("/user/updateUser", service.UpdateUser)
		//auth.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
		auth.POST("/user/find", service.FindByID)
		//发送消息
		auth.GET("/user/sendMsg", service.SendMsg)
		//发送消息
		auth.GET("/user/sendUserMsg", service.SendUserMsg)
		auth.POST("/searchFriends", service.SearchFriends)

		//添加好友
		auth.POST("/contact/addfriend", service.AddFriend)
		//上传文件
		auth.POST("/attach/upload", service.Upload)
		//创建群
		auth.POST("/contact/createCommunity", service.CreateCommunity)
		//群列表
		auth.POST("/contact/loadcommunity", service.LoadCommunity)
		auth.POST("/contact/joinGroup", service.JoinGroups)
		//心跳续命 不合适  因为Node  所以前端发过来的消息再receProc里面处理
		// r.POST("/user/heartbeat", service.Heartbeat)
		auth.POST("/user/redisMsg", service.RedisMsg)
	}

	return r
}
