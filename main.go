package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

type JsonResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

var c = cache.New(5*time.Minute, 10*time.Minute)

func HttpGetRequest(origin string, path string) (*http.Response, error) {
	resp, err := http.Get(origin + path)
	if err != nil {
		fmt.Printf("Error fetching origin %s: %s\n", origin, err)
		return nil, err
	}
	return resp, err
}
func HttpWriteResponse(writer http.ResponseWriter, resp JsonResponse) {
	writer.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		writer.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
	}
	writer.Write(jsonBytes)
}

func GetJsonResponse(r *http.Response) (JsonResponse, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return JsonResponse{}, err
	}
	jsonResp := JsonResponse{
		Headers:    r.Header,
		StatusCode: r.StatusCode,
		Body:       string(bodyBytes),
	}

	return jsonResp, nil
}

func CacheHandler(origin string, writer http.ResponseWriter, request *http.Request) {
	key := request.URL.Path
	cacheRes, found := c.Get(key)
	if found {
		cacheRes.(JsonResponse).Headers["X-Cache"] = []string{"HIT"}
		HttpWriteResponse(writer, cacheRes.(JsonResponse))
		log.Printf("Cache : HIT GET: %s\n", key)
		return
	}
	resp, err := HttpGetRequest(origin, key)
	if err != nil {
		log.Printf("Error fetching origin %s: %s\n", origin, err)
		return
	}
	jsonResp, err := GetJsonResponse(resp)
	if err != nil {
		log.Printf("Error parsing response: %s\n", err)
		return
	}
	jsonResp.Headers["X-Cache"] = []string{"MISS"}
	c.Set(key, jsonResp, cache.DefaultExpiration)
	HttpWriteResponse(writer, jsonResp)
	log.Printf("Cache : MISS GET: %s\n", key)
}

func GetArgSettings() (int, string) {
	port := flag.Int("port", 8080, "port to listen on")
	origin := flag.String("origin", "", "origin to allow")
	flag.Parse()

	return *port, *origin
}

func main() {
	port, origin := GetArgSettings()
	if origin == "" {
		log.Fatal("origin is required")
		return
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		CacheHandler(origin, w, r)
	})
	log.Printf("Listening on port %d, allowing origin %s", port, origin)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
