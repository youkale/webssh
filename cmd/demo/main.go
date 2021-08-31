package main

import "net/http"

func main() {
	http.ListenAndServe(":8082",http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello world"))
	}))
}
