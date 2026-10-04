package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Both modules use identical fixtures; the legacy service uses the v1 module
// compiled with GOEXPERIMENT=nojsonv2.
func benchmarkRequest(count int) Request {
	req := Request{Query: "search товары", Items: make([]Item, count)}
	for i := range req.Items {
		req.Items[i] = Item{
			ID: i, Name: fmt.Sprintf("item-%d товар", i), Price: float64(i) * 10.25,
			Tags: []string{"one", "two", "three"},
			Attributes: map[string]any{
				"category": "demo", "active": true, "score": float64(i) / 10,
				"description": strings.Repeat("details ", 8),
			},
		}
	}
	return req
}

var benchmarkEncoded []byte
var benchmarkDecoded Request

func BenchmarkJSON(b *testing.B) {
	for _, size := range []struct {
		name  string
		count int
	}{{"Small", 1}, {"Large", 1000}} {
		b.Run(size.name, func(b *testing.B) {
			req := benchmarkRequest(size.count)
			// Use the same v1 wire fixture in every variant, outside the timed loop.
			data, err := json.Marshal(req)
			if err != nil {
				b.Fatal(err)
			}
			resp := Response{Query: req.Query, ItemCount: len(req.Items), Items: req.Items}
			encoded, err := encodeJSON(resp)
			if err != nil {
				b.Fatal(err)
			}
			b.Run("Encode", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(encoded)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					out, err := encodeJSON(resp)
					if err != nil {
						b.Fatal(err)
					}
					benchmarkEncoded = out
				}
			})
			b.Run("Decode", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// Fresh destination, matching the HTTP handler: no allocation reuse.
					var dst Request
					if err := decodeJSON(data, &dst); err != nil {
						b.Fatal(err)
					}
					benchmarkDecoded = dst
				}
			})
		})
	}
}
