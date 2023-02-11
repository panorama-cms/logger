package logger

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var GeoIPDB *geoip2.Reader

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

	// Continent is the continent of the client.
	// Examples: Europe, North America, Asia, Africa, Oceania, Antarctica
	Continent string `json:"continent"`

	// Country is the country of the client.
	// Examples: Germany, United States, United Kingdom, France, Japan, China
	Country string `json:"country"`

	// CountryCode is the country code of the client.
	// Examples: DE, US, GB, FR, JP, CN
	CountryCode string `json:"country_code"`

	// City is the city of the client.
	// Examples: Berlin, New York, London, Paris, Tokyo, Shanghai
	City string `json:"city"`

	// Latitude is the latitude of the client.
	// Examples: 52.520008, 40.712776, 51.507351, 48.856613, 35.689487, 31.230416
	Latitude float64 `json:"latitude"`

	// Longitude is the longitude of the client.
	// Examples: 13.404954, -74.005974, -0.127758, 2.352222, 139.691706, 121.473701
	Longitude float64 `json:"longitude"`

	// Timezone is the timezone of the client.
	// Examples: Europe/Berlin, America/New_York, Europe/London, Europe/Paris, Asia/Tokyo, Asia/Shanghai
	Timezone string `json:"timezone"`

	// PostalCode is the postal code of the client.
	// Examples: 10115, 10001, SW1A 2AA, 75001, 100-0005, 200001
	PostalCode string `json:"postal_code"`

	// Subdivision is the subdivision of the client.
	// Examples: Berlin, New York, England, Île-de-France, Tokyo, Shanghai
	Subdivision string `json:"subdivision"`

	// SubdivisionCode is the subdivision code of the client.
	// Examples: BE, NY, ENG, IDF, 13, 31
	SubdivisionCode string `json:"subdivision_code"`
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
		"requested_host",
		"continent",
		"country",
		"country_code",
		"city",
		"latitude",
		"longitude",
		"timezone",
		"postal_code",
		"subdivision",
		"subdivision_code",
		"connection_id",
		"connection_seq",
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
		r.Continent + "," +
		r.Country + "," +
		r.CountryCode + "," +
		r.City + "," +
		fmt.Sprintf("%.12f", r.Latitude) + "," +
		fmt.Sprintf("%.12f", r.Longitude) + "," +
		r.Timezone + "," +
		r.PostalCode + "," +
		r.Subdivision + "," +
		r.SubdivisionCode + "," +
		strconv.FormatUint(r.ConnectionID, 10) + "," +
		strconv.FormatUint(r.ConnectionSeq, 10) + "\n"
}

func LogRequestFromFiber(c fiber.Ctx) {
	// Create a new request
	req := New()

	// Set the connection time
	connTime := time.Now().String()
	if c.Context() != nil {
		connTime = c.Context().ConnTime().String()
	}
	req.ConnectionTime = connTime

	// Set the method
	req.Method = c.Method()

	// Set the path
	req.Path = c.Path()

	// Set the IP
	var rawIP net.IP
	ip := c.IP()
	if len(c.IPs()) > 0 {
		ip = c.IPs()[0]
	}
	req.IP = ip
	rawIP = net.ParseIP(ip)

	if GeoIPDB != nil {
		record, err := GeoIPDB.City(rawIP)
		if err != nil {
			log.Fatal(err)
		}

		continent := "Unknown"
		if record.Continent.Names["en"] != "" {
			continent = record.Continent.Names["en"]
		}
		req.Continent = continent

		country := "Unknown"
		if record.Country.Names["en"] != "" {
			country = record.Country.Names["en"]
		}
		req.Country = country

		req.CountryCode = record.Country.IsoCode
		req.City = record.City.Names["en"]
		req.Latitude = record.Location.Latitude
		req.Longitude = record.Location.Longitude
		req.Timezone = record.Location.TimeZone
		req.PostalCode = record.Postal.Code

		subdivision := "Unknown"
		if len(record.Subdivisions) > 0 && record.Subdivisions[0].Names["en"] != "" {
			subdivision = record.Subdivisions[0].Names["en"]
		}
		req.Subdivision = subdivision

		subdivisionCode := "Unknown"
		if len(record.Subdivisions) > 0 && record.Subdivisions[0].IsoCode != "" {
			subdivisionCode = record.Subdivisions[0].IsoCode
		}
		req.SubdivisionCode = subdivisionCode
	}

	// Set the address
	remoteAddr := c.IP()
	if c.Context() != nil {
		remoteAddr = c.Context().RemoteAddr().String()
	}
	req.Address = remoteAddr

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

		filename := LogDir + "/requests-" + date + ".csv"

		// Add the header if the file doesn't exist
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// Create the file
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}

			// Write the header
			_, err = file.WriteString(strings.Join(GetCSVHeader(), ",") + "\n")
			if err != nil {
				log.Fatal(err)
			}
			err = file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}

		// open file requests.csv
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
