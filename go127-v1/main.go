package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

type Item struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Price      float64        `json:"price"`
	Tags       []string       `json:"tags"`
	Attributes map[string]any `json:"attributes"`
}
type Request struct {
	Query string `json:"query"`
	Items []Item `json:"items"`
}
type Response struct {
	Query     string `json:"query"`
	ItemCount int    `json:"item_count"`
	Items     []Item `json:"items"`
}

var requests, requestDuration, decodeDuration, encodeDuration uint64
var decodeAllocBytes, decodeMallocs, encodeAllocBytes, encodeMallocs uint64

func decodeJSON(data []byte, v *Request) error { return json.Unmarshal(data, v) }
func encodeJSON(v Response) ([]byte, error)    { return json.Marshal(v) }

func process(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	var req Request
	var decodeBefore, decodeAfter runtime.MemStats
	runtime.ReadMemStats(&decodeBefore)
	ds := time.Now()
	decodeErr := decodeJSON(body, &req)
	decodeElapsed := time.Since(ds)
	runtime.ReadMemStats(&decodeAfter)
	atomic.AddUint64(&decodeDuration, uint64(decodeElapsed.Nanoseconds()))
	atomic.AddUint64(&decodeAllocBytes, decodeAfter.TotalAlloc-decodeBefore.TotalAlloc)
	atomic.AddUint64(&decodeMallocs, decodeAfter.Mallocs-decodeBefore.Mallocs)
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), 400)
		return
	}

	resp := Response{Query: req.Query, ItemCount: len(req.Items), Items: req.Items}

	var encodeBefore, encodeAfter runtime.MemStats
	runtime.ReadMemStats(&encodeBefore)
	es := time.Now()
	out, err := encodeJSON(resp)
	encodeElapsed := time.Since(es)
	runtime.ReadMemStats(&encodeAfter)
	atomic.AddUint64(&encodeDuration, uint64(encodeElapsed.Nanoseconds()))
	atomic.AddUint64(&encodeAllocBytes, encodeAfter.TotalAlloc-encodeBefore.TotalAlloc)
	atomic.AddUint64(&encodeMallocs, encodeAfter.Mallocs-encodeBefore.Mallocs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
	atomic.AddUint64(&requests, 1)
	atomic.AddUint64(&requestDuration, uint64(time.Since(start).Nanoseconds()))
}

func metrics(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "# TYPE json_requests_total counter\njson_requests_total %d\n", atomic.LoadUint64(&requests))
	fmt.Fprintf(w, "# TYPE json_request_duration_seconds_sum counter\njson_request_duration_seconds_sum %f\n", float64(atomic.LoadUint64(&requestDuration))/1e9)
	fmt.Fprintf(w, "# TYPE json_decode_duration_seconds_sum counter\njson_decode_duration_seconds_sum %f\n", float64(atomic.LoadUint64(&decodeDuration))/1e9)
	fmt.Fprintf(w, "# TYPE json_encode_duration_seconds_sum counter\njson_encode_duration_seconds_sum %f\n", float64(atomic.LoadUint64(&encodeDuration))/1e9)
	fmt.Fprintf(w, "# TYPE json_decode_alloc_bytes_total counter\njson_decode_alloc_bytes_total %d\n", atomic.LoadUint64(&decodeAllocBytes))
	fmt.Fprintf(w, "# TYPE json_decode_mallocs_total counter\njson_decode_mallocs_total %d\n", atomic.LoadUint64(&decodeMallocs))
	fmt.Fprintf(w, "# TYPE json_encode_alloc_bytes_total counter\njson_encode_alloc_bytes_total %d\n", atomic.LoadUint64(&encodeAllocBytes))
	fmt.Fprintf(w, "# TYPE json_encode_mallocs_total counter\njson_encode_mallocs_total %d\n", atomic.LoadUint64(&encodeMallocs))
	fmt.Fprintf(w, "# TYPE go_goroutines gauge\ngo_goroutines %d\n", runtime.NumGoroutine())
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("ok")) })
	http.HandleFunc("/api/process", process)
	http.HandleFunc("/metrics", metrics)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
