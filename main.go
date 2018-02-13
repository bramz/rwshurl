package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/speps/go-hashids"
	"net/http"
	"time"
	"github.com/spf13/viper"
//	"regexp"
)


func main() {
	viper.SetConfigName("app")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	domain := viper.GetString("host")
	port := viper.GetString("port")
	dbuser := viper.GetString("dbuser")
	dbpass := viper.GetString("dbpass")
	dbname := viper.GetString("dbname")

	info := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbuser, dbpass, dbname) 
	db, err := sql.Open("postgres", info)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	fmt.Println("Starting application..,")

	router := gin.Default()

	router.LoadHTMLFiles("public/index.tmpl", "public/output.tmpl")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "public/index.tmpl", gin.H{
			"title": "rwshurl",
		})
	})

	router.POST("/s", func(c *gin.Context) {
		url := c.PostForm("url")

//		reg := regexp.MustCompile(`^(http|https):\/\/`)

		hashdata := hashids.NewData()
		hashdata.Salt = "pacifico is gay"
		hashdata.MinLength = 5

		h, _ := hashids.NewWithData(hashdata)

		now := time.Now()
		hash, _ := h.Encode([]int{int(now.Unix())})

		stmt := `INSERT INTO shortener (hash, url) VALUES ($1, $2)`
		_, err = db.Exec(stmt, hash, url)

		if err != nil {
			panic(err)
		}

		c.HTML(http.StatusOK, "public/output.tmpl", gin.H{
			"domain": domain,
			"title": "rwshurl output",
			"hash":  hash,
			"url":   url,
		})

	})

	router.GET("/s/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		stmt := `SELECT url from shortener where hash=$1`

		row := db.QueryRow(stmt, hash)
		switch err := row.Scan(&hash); err {
		case sql.ErrNoRows:
			fmt.Println("No rows returned!")
		case nil:
			c.Redirect(http.StatusMovedPermanently, hash)
		default:
			panic(err)
		}
	})

	router.Run(":" + port)
}
