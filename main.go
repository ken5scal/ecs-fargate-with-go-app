package main
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"golang.org/x/xerrors"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"encoding/json"
)

type TestLog struct {
	Status int `json:"status"`
	Result string `json:"result"`
	ErrorResult struct {
	   ErrorCode int `json:"errorCode"`
	   Reason string `json:"reason"`
	} `json:"errorResult"`
}

func main() {
	testLog := TestLog{}

	arg := os.Args[1]
	i, _ := strconv.Atoi(arg)
	fmt.Println("here is arg: " + arg)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
	m := http.NewServeMux()
	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) { 
		message := "/shutdown called and starting graceful shutdown in "
		
		testLog.Status = http.StatusOK
		testLog.Result = fmt.Sprintf(message + "%vsec", i)
		
		res, _ := json.Marshal(testLog)

		fmt.Println(string(res))
		w.WriteHeader(http.StatusOK)
		w.Write(res)
		sigCh <- syscall.SIGTERM
	})

	m.HandleFunc("/std", func(w http.ResponseWriter, r *http.Request) {
		testLog.Status = http.StatusOK
		testLog.Result = "ok"

		res, _ := json.Marshal(testLog)

		fmt.Println(string(res))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	m.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		e1 := xerrors.New("error using /x/xerrors package")
		testLog.Status = http.StatusInternalServerError
		testLog.Result = e1.Error()

		testLog = TestLog {
			Status: http.StatusInternalServerError,
			Result: "",
			ErrorResult: struct{
				ErrorCode int  `json:"errorCode"`
				Reason string   `json:"reason"`
			}{
				ErrorCode: http.StatusInternalServerError,
				Reason: e1.Error(),
			},
		}

		res, _ := json.Marshal(testLog)

		fmt.Fprintln(os.Stderr, string(res))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(res)
	})

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hellow world")
	})
	srv := http.Server{Addr: "0.0.0.0:80", Handler: m}
	go func() {
		<-sigCh
		fmt.Println("SIGTERM Received")
		fmt.Println("some left over transactions before shutting down")
		time.Sleep(time.Duration(i) * time.Second)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Shutting down go app: %v", err)
		} else {
			fmt.Println("Golang App Terminated")
		}
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Print(err)
	}
}
