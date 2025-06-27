package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	mymw "github.com/aditi2420/fleet-tracker/internal/middleware"
)

func main() {
	var (
		sub    = flag.String("sub", "dev", "subject/user id")
		ttl    = flag.Duration("exp", 24*time.Hour, "token validity")
		secret = flag.String("secret", "", "signing secret (env JWT_SIGN_KEY overrides)")
	)
	flag.Parse()

	sec := []byte(*secret)
	if len(sec) == 0 {
		sec = []byte(os.Getenv("JWT_SIGN_KEY"))
	}
	if len(sec) == 0 {
		log.Fatal("provide -secret or set JWT_SIGN_KEY")
	}

	tok, err := mymw.GenerateDevToken(*sub, sec, *ttl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tok)
}
