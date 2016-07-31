package common

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

//Global ID tracking for Commands
//Could probably move this to a factory function
var ActionId uint = 0
var LogLevel string = "enabled"
var Log *log.Logger = log.New(ioutil.Discard, "Go-Thrust:", 3)

func InitLogger(level string) {
	LogLevel = strings.ToLower(level)
	switch LogLevel {
	case "none":
		Log = log.New(ioutil.Discard, "Go-Thrust:", 3)
	default:
		Log = log.New(os.Stdout, "Go-Thrust:", 3)
	}
	Log.Print("Thrust Client:: Initializing")
}
