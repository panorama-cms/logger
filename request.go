package logger

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	// ConnectionTime is the connection time of the client.
	// See https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.ConnTime
	ConnectionTime string `json:"connection_time"`

	// Method is the HTTP method used for the request.
	// Examples: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
	Method string `json:"method"`

	// Path is the path of the request.
	// Examples: /, /api, /api/v1, /api/v1/endpoint
	Path string `json:"path"`

	// IP is the IP address of the client. This may be a IP v4 or IP v6 address.
	// Examples: 127.0.0.1, ::1
	IP string `json:"ip"`

	// Address is the address of the client.
	// Examples: localhost, 127.0.0.1, ::1
	Address string `json:"address"`

	// UserAgent is the user agent of the client.
	// Examples: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36
	UserAgent string `json:"user_agent"`

	// Referer is the referer of the client.
	// Examples: https://google.com, https://some.host/some/path
	Referer string `json:"referer"`

	// ConnectionID is the connection ID of the client.
	// See https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.ConnID
	ConnectionID uint64 `json:"connection_id"`

	// ConnectionSeq is the connection sequence of the client.
	// See https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.ConnRequestNum
	ConnectionSeq uint64 `json:"connection_seq"`

	// RequestedHost is the requested host of the client.
	// See https://pkg.go.dev/github.com/valyala/fasthttp#RequestCtx.Host
	RequestedHost string `json:"requested_host"`
}

func New() *Request {
	return &Request{}
}

func (r *Request) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func GetCSVHeader() []string {
	return []string{
		"connection_time",
		"method",
		"path",
		"ip",
		"address",
		"user_agent",
		"referer",
		"connection_id",
		"connection_seq",
		"requested_host",
	}
}

func (r *Request) ToCSV() string {
	return r.ConnectionTime + "," +
		r.Method + "," +
		r.Path + "," +
		r.IP + "," +
		r.Address + "," +
		r.UserAgent + "," +
		r.Referer + "," +
		r.RequestedHost + "," +
		strconv.FormatUint(r.ConnectionID, 10) + "," +
		strconv.FormatUint(r.ConnectionSeq, 10) + "\n"
}

func LogRequestFromFiber(c *fiber.Ctx) {
	// Create a new request
	req := New()

	// Set the connection time
	req.ConnectionTime = c.Context().ConnTime().String()

	// Set the method
	req.Method = c.Method()

	// Set the path
	req.Path = c.Path()

	// Set the IP
	req.IP = c.IP()

	// Set the address
	req.Address = c.Context().RemoteAddr().String()

	// Set the user agent
	req.UserAgent = c.Get(fiber.HeaderUserAgent)

	// Set the referer
	req.Referer = c.Get(fiber.HeaderReferer)

	// Set the connection ID
	req.ConnectionID = c.Context().ConnID()

	// Set the connection sequence
	req.ConnectionSeq = c.Context().ConnRequestNum()

	// Set the requested host
	req.RequestedHost = string(c.Context().Host())

	// Log the request
	LogRequest(req)
}

func LogRequest(req *Request) {
	if (!LogRequestsSeparately) || (LogRequestsSeparately && !HideRequestsFromMainLog) {
		Log(LevelInfo, fmt.Sprintf("(%s) %s <- %s @ %s", req.Method, req.Path, req.UserAgent, req.IP))
	}

	if LogRequestsSeparately {
		// get the current date
		t := time.Now()

		// format time to YYYY-MM-DD
		date := t.Format("2006-01-02")

		// format time to HH:MM:SS
		//tFormatted := t.Format("2006-01-02 15:04:05.000000")

		// open file requests.csv
		f, err := os.OpenFile(LogDir+"/requests-"+date+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// replace all , with ; in user agent
		req.UserAgent = strings.ReplaceAll(req.UserAgent, ",", ";")

		entry := req.ToCSV()

		// write to file
		_, err = f.WriteString(entry)
		if err != nil {
			log.Fatal(err)
		}
	}
}