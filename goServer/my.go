package main

/*
import (
    "html/template"
    "log"
    "net/http"
    "time"
    "fmt"
    "math/rand"
    "github.com/joho/godotenv"
    "os"
)
*/

import (
    "log"
    "math/rand"
    "github.com/joho/godotenv"
    "os"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func goDotEnvVariable(key string) string {

  // load .env file
  err := godotenv.Load(".env")

  if err != nil {
    log.Fatalf("Error loading .env file")
  }

  return os.Getenv(key)
}