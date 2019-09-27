package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
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
	conf   *Config
)

var (
	flag_v, flag_debug, flag_help bool
)

type Config struct {
	Debug      bool
	GOMAXPROCS int
	Authen     bool
	Http_port  string
	User_token []string
	Allow_ip   []string
}

func init() {
	fmt.Println("init.....")

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

	if conf.Debug {
		fmt.Println(conf)
	}

}

func PingHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(r.RemoteAddr + "\nAlita(小艾):I'm here!\n"))
}

func main() {
	if conf.GOMAXPROCS > 0 {
		runtime.GOMAXPROCS(conf.GOMAXPROCS)
	}

	logger.Println("Start......")

	//Single instance
	Singleton()

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
	/*
		r.HandleFunc("/posts", getPosts).Methods("GET")
		r.HandleFunc("/posts", createPost).Methods("POST")
		r.HandleFunc("/posts/{id}", getPost).Methods("GET")
		r.HandleFunc("/posts/{id}", updatePost).Methods("PUT")
		r.HandleFunc("/posts/{id}", deletePost).Methods("DELETE")
	*/
	r.HandleFunc("/", PingHandler).Methods("GET")
	r.HandleFunc("/log", LogHandler).Methods("GET")
	r.HandleFunc("/simple", SimpleHandler).Methods("GET")
	r.HandleFunc("/host/info", getHostInfoHandler).Methods("GET")
	r.HandleFunc("/host/name", updateHostInfoHandler).Methods("PUT")

	logger.Fatal(http.ListenAndServe(":"+conf.Http_port, r))
}
