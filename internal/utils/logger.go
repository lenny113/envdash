package utils

import (
	"log"
	"net/http"
	"os"
	"time"
)

// Creating a  dedicated logging instance for logging http requests
var HttpLogger *log.Logger

// Function for initializing the logger, telling it to log to the "requsts.log" file that
// will either be created id it does not exist and appended to if it does exist.
func InitLogger() {
	file, err := os.OpenFile("requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	// We set the flags to 0 because we are formatting the date/time manually in the table
	HttpLogger = log.New(file, "", 0)

	// Print table heads to the file once with correct spacing
	HttpLogger.Printf("%-10s %-8s %-6s %-8s %-25s %-10s %s",
		"Date", "Time", "Status", "Method", "Path", "Duration", "Message")
}

// wrappedWriter embeds http.ResponseWriter to "intercept" and store the status code
// and custom messages for logging purposes.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
	message    string
}

// WriteHeader captures the status code before passing it to the ResponseWriter again.
func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Function to set messages for a request or response
// usage:
// SetMessageForLogger(<responsewriter variable>, <message to add to log for the given request>)
func SetMessageForLogger(w http.ResponseWriter, msg string) {
	if ww, ok := w.(*wrappedWriter); ok {
		ww.message = msg
	}
}

// Logging is a middleware that records the details of HTTP requests being made.
// It also calculates the duration of the http requests and adds it to the records.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//Initialize wrappwer with default 200 OK status
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		//process request
		next.ServeHTTP(wrapped, r)

		//calculate latatency and capture time after completion
		duration := time.Since(start) // duration of request
		now := time.Now()

		// Write the formatted entry to the "requests.log" file
		HttpLogger.Printf("%-10s %-8s %-6d %-8s %-25s %-10v %s",
			now.Format("2006-01-02"),
			now.Format("15:04:05"),
			wrapped.statusCode,
			r.Method,
			r.URL.Path,
			duration,
			wrapped.message,
		)
	})
}
