package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yahoojapan/yisucon/benchmarker/config"
	"github.com/yahoojapan/yisucon/benchmarker/logger"
	"github.com/yahoojapan/yisucon/benchmarker/runner"
)

func main() {

	l := logger.GetLogger()

	defer func() {
		if err := recover(); err != nil {
			l.Println(err)
		}
		log.Fatalln(l.Close())
	}()

	if strings.Contains(config.PortalHost, "localhost") || config.PortalHost == "" {
		l.Fatalln(errors.New("Invalide PortalHost"))
	}

	_, err := http.DefaultClient.Get(fmt.Sprintf("http://%s", config.PortalHost))

	if err != nil {
		l.Fatalln(err)
	}

	l.Fatalln(runner.Run())
}
