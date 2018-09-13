package main

import (
	"flag"
	"fmt"
	"os"

	redirector "github.com/rmanzoku/s3-path-redirector"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		os.Exit(1)
	}

	subcmd := args[0]
	key := args[1]

	switch subcmd {
	case "get":
		ret, err := get(key)
		if err != nil {
			panic(err)
		}
		fmt.Println(ret)

	default:
		fmt.Println("Subcmd must be get")
		os.Exit(1)
	}
}

func get(key string) (string, error) {
	r, _ := redirector.NewRedirector()
	r.Region = os.Getenv("AWS_REGION")
	r.Bucket = os.Getenv("S3_BUCKET")
	r.Prepare()

	r.RedirectToFormat = "https://www.mycryptoheroes.net/?s=%s"
	return r.CreateLink(key)
}
