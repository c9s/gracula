package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/linkernetworks/logger"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Arithmetic struct{}

type MultiplyArgs struct {
	A, B int
}

type MultiplyResult struct {
	Result int
}

func (a *Arithmetic) Multiply(args *MultiplyArgs, reply *MultiplyResult) error {
	reply.Result = args.A * args.B
	return nil
}

var ()

func main() {
	var bind string
	var verbose bool
	kingpin.Flag("verbose", "Verbose mode.").Short('v').BoolVar(&verbose)
	kingpin.Arg("bind", "bind address").Default(":9702").StringVar(&bind)
	kingpin.Parse()

	logger.Setup(logger.LoggerConfig{})

	s := rpc.NewServer()

	arithmetic := &Arithmetic{}
	s.Register(arithmetic)

	http.HandleFunc("/jrpc", func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			log.Fatalln(err)
		}
		s.ServeCodec(jsonrpc.NewServerCodec(conn))
	})

	logger.Infof("Listening at %s", bind)
	http.ListenAndServe(bind, nil)
}

func client() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	io.WriteString(conn, "CONNECT "+"/jrpc"+" HTTP/1.0\n\n")
	if err != nil {
		log.Fatalln(err)
	}

	cli := jsonrpc.NewClient(conn)
	var ret MultiplyResult
	err = cli.Call("Arithmetic.Multiply", &MultiplyArgs{A: 5, B: 5}, &ret)
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()
	log.Println(ret.Result)
}
