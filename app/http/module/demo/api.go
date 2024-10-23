package demo

import (
	"database/sql"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/gin"
	"github.com/Superdanda/hade/framework/provider/orm"
	"time"
)

type DemoApi struct {
	service *Service
}

func Register(r *gin.Engine) error {
	api := NewDemoApi()
	//r.Bind(&demoService.DemoProvider{})

	r.GET("/demo/demo", api.Demo)
	r.GET("/demo/demo2", api.Demo2)
	r.GET("/demo/orm", api.orm)
	r.POST("/demo/demo_post", api.DemoPost)
	return nil
}

func NewDemoApi() *DemoApi {
	service := NewService()
	return &DemoApi{service: service}
}

func Demo(c *gin.Context) {
	configService := c.MustMake(contract.ConfigKey).(contract.Config)
	log := c.MustMake(contract.LogKey).(contract.Log)
	password := configService.GetString("database.mysql.password")
	log.Info(c, "ceshiceshi", map[string]interface{}{})
	c.JSON(200, password+"后端测试222")
}

// Demo godoc
// @Summary 获取所有用户
// @tag.description.markdown demo.md
// @Produce  json
// @Tags demo
// @Success 200 array []UserDTO
// @Router /demo/demo [get]
func (api *DemoApi) Demo(c *gin.Context) {
	c.JSON(200, "this is demo for dev all")
}

func (api *DemoApi) orm(c *gin.Context) {
	logger := c.MustMakeLog()
	logger.Info(c, "request start", nil)

	// 初始化一个orm.DB
	gormService := c.MustMake(contract.ORMKey).(contract.ORMService)
	db, err := gormService.GetDB(orm.WithConfigPath("database.default"))
	if err != nil {
		logger.Error(c, err.Error(), nil)
		c.AbortWithError(50001, err)
		return
	}
	db.WithContext(c)

	// 将User模型创建到数据库中
	err = db.AutoMigrate(&User{})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	logger.Info(c, "migrate ok", nil)

	// 插入一条数据
	email := "foo@gmail.com"
	name := "foo"
	age := uint8(25)
	birthday := time.Date(2001, 1, 1, 1, 1, 1, 1, time.Local)
	user := &User{
		Name:         name,
		Email:        &email,
		Age:          age,
		Birthday:     &birthday,
		MemberNumber: sql.NullString{},
		ActivatedAt:  sql.NullTime{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(user).Error
	logger.Info(c, "insert user", map[string]interface{}{
		"id":  user.ID,
		"err": err,
	})

	// 更新一条数据
	user.Name = "bar"
	err = db.Save(user).Error
	logger.Info(c, "update user", map[string]interface{}{
		"err": err,
		"id":  user.ID,
	})

	// 查询一条数据
	queryUser := &User{ID: user.ID}

	err = db.First(queryUser).Error
	logger.Info(c, "query user", map[string]interface{}{
		"err":  err,
		"name": queryUser.Name,
	})

	// 删除一条数据
	//err = db.Delete(queryUser).Error
	//logger.Info(c, "delete user", map[string]interface{}{
	//	"err": err,
	//	"id":  user.ID,
	//})
	c.JSON(200, "ok")
}

// Demo2  for godoc
// @Summary 获取所有学生
// @Description 获取所有学生,不进行分页
// @Produce  json
// @Tags demo
// @Success 200 {array} UserDTO
// @Router /demo/demo2 [get]
func (api *DemoApi) Demo2(c *gin.Context) {
	//demoProvider := c.MustMake(demoService.DemoKey).(demoService.IService)
	//students := demoProvider.GetAllStudent()
	//usersDTO := StudentsToUserDTOs(students)
	c.JSON(200, "usersDTO")
}

func (api *DemoApi) DemoPost(c *gin.Context) {
	type Foo struct {
		Name string
	}
	foo := &Foo{}
	err := c.BindJSON(&foo)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.JSON(200, nil)
}
