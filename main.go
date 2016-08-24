package main

import "github.com/gin-gonic/gin"
import "encoding/json"
import "os"
import "log"
import "github.com/DreamHacks/sc2a-service/dist"
import "net/http"
import "errors"
import "text/template"

// Config is the app config
type Config struct {
	BaseURI  string
	User     string
	Password string
	Dist     dist.Config
}

var config Config
var nameTemplate *template.Template

func init() {
	f, err := os.Open("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dist.Configure(config.BaseURI, config.Dist)
	dist.OpenDB()
	defer dist.CloseDB()

	r := gin.Default()

	api := r.Group("/api")

	api.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			if c.Writer.Status() == http.StatusOK {
				c.Status(http.StatusInternalServerError)
			}
			errors := c.Errors
			c.JSON(-1, gin.H{"message": errors[0].Error(), "errors": errors})
		}
	})

	authAccounts := gin.Accounts{}
	authAccounts[config.User] = config.Password
	api.Use(gin.BasicAuth(authAccounts))

	api.GET("/release", func(c *gin.Context) {
		list, err := dist.List()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, list)
	})

	api.POST("/release", func(c *gin.Context) {
		req := c.Request
		r := dist.Release{
			Version:     req.FormValue("Version"),
			Description: req.FormValue("Description"),
		}
		if r.Version == "" {
			c.Status(http.StatusBadRequest)
			c.Error(errors.New("Version is required"))
			return
		}
		f, _, err := req.FormFile("File")
		if err != nil {
			c.Status(http.StatusBadRequest)
			if err == http.ErrMissingFile {
				c.Error(errors.New("Please upload a file"))
			} else {
				c.Error(err)
			}
			return
		}

		published, err := dist.Publish(r, f)
		if err != nil {
			c.Error(err)
			return
		}

		err = dist.NotifyAll(*published)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, published)
	})

	api.GET("/release/:id", func(c *gin.Context) {
		id := c.Param("id")
		r, err := dist.Get(id)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, r)
	})

	api.DELETE("/release/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := dist.Unpublish(id)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	api.GET("/sub", func(c *gin.Context) {
		list, err := dist.ListSubs()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, list)
	})

	api.POST("/sub", func(c *gin.Context) {
		sub := dist.Sub{}
		err := c.BindJSON(&sub)
		if err != nil {
			c.Status(http.StatusBadRequest)
			c.Error(err)
			return
		}
		created, err := dist.Subscribe(sub)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, created)
	})

	api.PUT("/sub", func(c *gin.Context) {
		sub := dist.Sub{}
		err := c.BindJSON(&sub)
		if err != nil {
			c.Status(http.StatusBadRequest)
			c.Error(err)
			return
		}
		updated, err := dist.UpdateSubscriber(sub)
		if err != nil {
			if err == dist.ErrSubNotFound {
				c.Status(http.StatusBadRequest)
			}
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, updated)
	})

	api.DELETE("/sub/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := dist.Unsubscribe(id)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	api.GET("/sub/:id/stats", func(c *gin.Context) {
		id := c.Param("id")
		stats, err := dist.GetSubDowloadStats(id)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, stats)
	})

	r.GET("/download/:id", func(c *gin.Context) {
		id := c.Param("id")
		link, err := dist.GetLink(id)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		release, err := dist.Get(link.ReleaseID)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if release == nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.Header("Content-Disposition", "attachment; filename="+release.FileName())
		c.Header("Content-Type", "application/octet-stream")
		err = dist.StreamLink(id, c.Writer)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	})

	r.Run()
}
