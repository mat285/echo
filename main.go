package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.All()

	appStart := time.Now()

	contents, err := ioutil.ReadFile(env.Env().String("CONFIG_PATH", "/var/secrets/config.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
	})
	app.GET("/text/:text", func(r *web.Ctx) web.Result {
		return r.Text().Result(r.Param("text"))
	})
	app.GET("/headers", func(r *web.Ctx) web.Result {
		contents, err := json.Marshal(r.Request.Header)
		if err != nil {
			return r.View().InternalError(err)
		}
		return r.Text().Result(string(contents))
	})
	app.GET("/env", func(r *web.Ctx) web.Result {
		return r.JSON().Result(env.Env().Vars())
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 12*time.Second {
			return r.Text().Result("OK!")
		}
		return r.Text().BadRequest("not ready")
	})
	app.GET("/config", func(r *web.Ctx) web.Result {
		r.Response.Header().Set("Content-Type", "application/yaml") // but is it really?
		return r.Raw(contents)
	})
	app.GET("/long", func(r *web.Ctx) web.Result {
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				{
					fmt.Fprintf(r.Response, "tick\n")
					r.Response.Flush()
				}
			}
		}

		return nil
	})
	app.GET("/echo/*filepath", func(r *web.Ctx) web.Result {
		body := r.Request.URL.Path
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("no response."))
		}
		return r.RawWithContentType(web.ContentTypeText, []byte(body))
	})
	app.POST("/echo/*filepath", func(r *web.Ctx) web.Result {
		body, err := r.PostBody()
		if err != nil {
			return r.JSON().InternalError(err)
		}
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("nada."))
		}
		return r.RawWithContentType(web.ContentTypeText, body)
	})

	go ping()

	log.Fatal(app.Start())
}

func ping() {
	url := env.Env().String("PING_URL")
	if len(url) == 0 {
		fmt.Println("No url for pinging")
		return
	}
	for {
		time.Sleep(time.Second)
		key := make([]byte, 12)
		rand.Read(key)
		rurl := url + "/text/hello%20my%20secret%20is%20" + hex.EncodeToString(key)
		req, err := http.NewRequest("GET", rurl, nil)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(string(data))
	}
}
