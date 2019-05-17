package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	VERSION    = "0.1"
	EVENTS_LOG = "events.log"
	ERROR_LOG  = "error.log"
)

var (
	errLog *log.Logger
	logger *log.Logger
	conf   Config
)

var (
	flag_v, flag_debug, flag_help bool
)

type Config struct {
	Debug  bool
	Authen bool
}

func init() {
	flag.BoolVar(&flag_v, "v", false, "Version")
	flag.BoolVar(&flag_debug, "debug", false, "Debug")
	flag.BoolVar(&flag_help, "h", false, "Help")
	flag.Parse()

	if flag_v {
		fmt.Println("Version:", VERSION)
		os.Exit(0)
	}

	if flag_help {
		flag.Usage()
		os.Exit(0)
	}

	//create error logger
	f0, err := os.OpenFile(ERROR_LOG, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		errLog.Fatal("[error]opening error file: %v", err)
	}
	errLog = log.New(f0, "", log.Lshortfile|log.LstdFlags)

	//create events logger
	f1, err := os.OpenFile(EVENTS_LOG, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		errLog.Fatal("[error]opening error file: %v", err)
	}
	logger = log.New(f1, "", log.LstdFlags)

	//Decode config.toml
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		errLog.Fatal(err)
	}

}

func PingHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(r.RemoteAddr + "\nAlita(小艾):I'm here!\n"))
}

func main() {
	runtime.GOMAXPROCS(2)
	logger.Println("Start......")

	//Single instance
	Singleton()

	//connect to DB
	//db, err := DBConnect()
	//if err != nil {
	//	os.Exit(0)
	//}
	//db.Close()

	//Create Router
	r := mux.NewRouter()

	//use logging Middleware
	r.Use(loggingMiddleware)

	///use auth Middleware
	amw := authenticationMiddleware{}
	amw.Populate()
	if conf.Authen {
		r.Use(amw.Middleware)
	}

	//handler
	r.HandleFunc("/", PingHandler)
	r.HandleFunc("/log", LogHandler)
	r.HandleFunc("/simple", SimpleHandler)
	r.HandleFunc("/host/info", getHostInfoHandler)
	//r.HandleFunc("/host/setname", HostInfoHandler)

	logger.Fatal(http.ListenAndServe(":9900", r))
}
