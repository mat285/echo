package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.NewFromEnvironment()

	appStart := time.Now()

	awscreds, err := ioutil.ReadFile("/var/aws-credentials/config")
	if err != nil {
		fmt.Println(err)
		time.Sleep(time.Minute * 3)
		os.Exit(1)
	}

	parts := strings.Split(string(awscreds), "\n")
	var access, secret string
	for _, part := range parts {
		san := strings.TrimSpace(part)
		if strings.HasPrefix(san, "aws_access_key_id=") {
			access = strings.TrimSpace(strings.TrimPrefix(san, "aws_access_key_id="))
		}
		if strings.HasPrefix(san, "aws_secret_access_key=") {
			secret = strings.TrimSpace(strings.TrimPrefix(san, "aws_secret_access_key="))
		}
	}

	if len(access) == 0 || len(secret) == 0 {
		fmt.Println("Missing some creds")
		fmt.Println(access)
		fmt.Println(secret)
		os.Exit(1)
	}
	data := strings.NewReader("hello")
	input := &s3.PutObjectInput{
		Bucket: aws.String("blend-testing-creds"),
		Key:    aws.String("key"),
		Body:   data,
	}

	serv := s3.New(session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(access, secret, ""),
	}))

	_, err = serv.PutObject(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(env.Env().String("CONFIG_PATH", "/var/secrets/config.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
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

	log.Fatal(app.Start())
}
