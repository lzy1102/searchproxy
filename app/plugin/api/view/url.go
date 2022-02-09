package view

import (
	"github.com/gin-gonic/gin"
	"searchproxy/app/plugin/api/view/proxyview"
)

func Route(app *gin.Engine) {
	api := app.Group("api")
	{
		get := api.Group("get")
		{
			get.GET("list", proxyview.GetProxylist)
		}
		post := api.Group("post")
		{
			post.POST("delete", proxyview.DeleteProxy)
			post.POST("addport", proxyview.AddPort)
		}
	}
}
