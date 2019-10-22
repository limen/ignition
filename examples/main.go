package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/limen/ignition"
	"github.com/limen/ignition/middlewares"
	"github.com/limen/ignition/validation"
	"strings"
)

type tzGetter struct{}
type localeGetter struct{}
type configParamBag struct{}

// user post entity which holds
// - post data
// - validation rules
// - validation errors (if any)
type UserPostEntity struct {
	ignition.RequestEntity
}

// user post data structure
type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// define user schema
// id - int primary key
// username - varchar
// password - varchar
type UserModelEntity struct {
	ignition.ModelEntity
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// user CRUD functions container
type UserModel struct {
	ignition.Model
}

// environment configuration
type config struct {
	Timezone   string `yml:"timezone"`
	Locale     string `yml:"locale"`
	DbHost     string `yml:"dbhost"`
	DbName     string `yml:"dbname"`
	DbPort     int    `yml:"dbport"`
	DbUser     string `yml:"dbuser"`
	DbPassword string `yml:"dbpassword"`
}

// postgres connection pool
var pgPool = &ignition.Pool{
	MaxActive: 10,
	MaxIdle:   1,
	Dial: func() (ignition.Conn, error) {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			conf.DbHost,
			conf.DbPort,
			conf.DbUser,
			conf.DbName,
			conf.DbPassword,
			"disable",
		)
		db, err := gorm.Open("postgres", dsn)
		if err != nil {
			panic("Database connection failed:" + err.Error())
		}

		return db, db.Error
	},
}

var conf config
var confLoader ignition.Config

// getters
var TzGetter = tzGetter{}
var LocaleGetter = localeGetter{}

// getter map
var getterMap = map[string]ignition.Getter{
	"tz":     TzGetter,
	"locale": LocaleGetter,
}

func newUserPostEntity(ctx *gin.Context) UserPostEntity {
	e := UserPostEntity{}
	// validation rules
	e.Rules = validation.Rules{
		"username": validation.Username,
		"password": validation.Password,
	}
	// post data
	e.Data = UserData{
		Username: ctx.PostForm("username"),
		Password: ctx.PostForm("password"),
	}

	return e
}

func (en *UserPostEntity) Validate() {
	en.RequestEntity.Validate()
	// customize validation logic here
	if strings.ContainsAny(en.Data.(UserData).Username, "_") {
		en.AddError("username", "username shouldn't contain '_'")
	}
}

// user schema table name
func (UserModelEntity) TableName() string {
	return "ignitor_users"
}

func NewUserModel() *UserModel {
	return &UserModel{}
}

// create user
func (m *UserModel) Create(username string, password string) (uint, error) {
	db, err := pgPool.Get()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	user := UserModelEntity{
		Username: username,
		Password: password,
	}
	userErr := db.GetDB().Create(&user).Error

	return user.ID, userErr
}

// find user
func (m *UserModel) Find(username string) (UserModelEntity, error) {
	db, err := pgPool.Get()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	user := UserModelEntity{}
	dbErr := db.GetDB().Where("username=?", username).First(&user).Error
	return user, dbErr
}

// getter business logic
func (tzGetter) Get(ctx interface{}, allGetters map[string]bool) interface{} {
	return conf.Timezone
}

func (localeGetter) Get(ctx interface{}, allGetters map[string]bool) interface{} {
	return conf.Locale
}

func main() {
	r := gin.New()
	// load yaml configuration file
	confLoader.Load(".env.yml", &conf)

	// use middleware
	r.Use(middlewares.AccessLogHandler())
	r.Use(middlewares.PanicLogger())

	// get user info
	r.GET("/user-info", func(ctx *gin.Context) {
		data := map[string]interface{}{}
		user, err := NewUserModel().Find(ctx.Query("username"))
		data["user"] = user
		data["error"] = err
		ignition.Response.Success(ctx, data)
	})
	// create user
	r.POST("/users", func(ctx *gin.Context) {
		// create post entity
		entity := newUserPostEntity(ctx)
		// validate post data
		entity.Validate()
		// if there're errors
		// response with error code
		if entity.HaveErrors() {
			ignition.Response.Error(ctx, "ParamError", "Param validation failed", entity.Errors)
		} else {
			userData := entity.Data.(UserData)
			_, err := NewUserModel().Create(userData.Username, userData.Password)
			if err != nil {
				ignition.Response.Error(ctx, "DatabaseError", "Create user error:"+err.Error(), nil)
			} else {
				ignition.Response.Success(ctx, map[string]interface{}{"username": userData.Username})
			}
		}
	})
	// use getters to enable clients to get what they need
	r.GET("/config", func(ctx *gin.Context) {
		data := map[string]interface{}{}
		confParamBag := &configParamBag{}
		query := strings.Split(ctx.Query("getters"), ",")
		getters := map[string]bool{}
		for _, q := range query {
			getters[q] = true
		}
		for getter := range getters {
			if !getters[getter] {
				continue
			}
			data[getter] = getterMap[getter].Get(confParamBag, getters)
		}

		ignition.Response.Success(ctx, data)
	})
	// see formatted panic in stdout
	r.GET("/panic", func(ctx *gin.Context) {
		panic("what's wrong, buddy?")
	})
	// listen port 8765
	r.Run(":8765")
}
