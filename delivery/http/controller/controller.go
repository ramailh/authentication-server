package controller

import (
	"log"
	"net/http"

	"github.com/ramailh/authentication-server/delivery/http/middleware"
	"github.com/ramailh/authentication-server/models"
	"github.com/ramailh/authentication-server/service"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service service.Services
}

func NewUserRoute(router *gin.Engine, srv service.Services) {
	cnt := controller{service: srv}

	router.POST("/user/", cnt.Register)

	login := router.Group("/login")
	{
		login.POST("/", cnt.Login)
		login.GET("/google", cnt.Google)
		login.GET("/call-back", cnt.GoogleCallback)
	}

	user := router.Group("/user", middleware.VerifyToken)
	{
		user.GET("/", cnt.FindAll)
		user.GET("/:id", cnt.FindByID)
		user.PUT("/:id", cnt.Update)
		user.DELETE("/:id", cnt.Delete)
	}
}

func (cnt *controller) Register(c *gin.Context) {
	var param models.User
	if err := c.BindJSON(&param); err != nil {
		log.Println(err)
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	data, err := cnt.service.Register(param)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) Login(c *gin.Context) {
	var param models.User
	if err := c.BindJSON(&param); err != nil {
		log.Println(err)
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	data, err := cnt.service.Login(param.Username, param.Password)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) Google(c *gin.Context) {
	url := cnt.service.Google()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (cnt *controller) GoogleCallback(c *gin.Context) {
	code := c.Query("code")

	data, err := cnt.service.GoogleCallback(code)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) FindAll(c *gin.Context) {
	var param models.GetAll
	if err := c.Bind(&param); err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	data, err := cnt.service.FindAll(param.SortType, param.SortBy, param.WID, param.Search, param.From, param.Limit)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) FindByID(c *gin.Context) {
	id := c.Param("id")

	data, err := cnt.service.FindByID(id)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) Update(c *gin.Context) {
	var param models.User
	if err := c.BindJSON(&param); err != nil {
		log.Println(err)
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}
	param.ID = c.Param("id")

	data, err := cnt.service.Update(param)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}

func (cnt *controller) Delete(c *gin.Context) {
	id := c.Param("id")

	data, err := cnt.service.Delete(id)
	if err != nil {
		c.JSON(200, gin.H{"status": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "success", "data": data})
}
