package main

import (
	"net"
	"net/http"
	"strings"
)

//type MiddlewareFunc func(http.Handler) http.Handler

// Define auth struct
type authenticationMiddleware struct {
	tokenUsers map[string]string
	allowIPs   map[string]string
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		logger.Println(r.RemoteAddr, r.Method, r.Referer(), r.RequestURI)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// Initialize authenticationMiddleware
func (amw *authenticationMiddleware) Populate() {
	amw.tokenUsers = make(map[string]string)
	amw.allowIPs = make(map[string]string)

	//Populate token
	//amw.tokenUsers["00000000"] = "user0"
	for _, t := range conf.User_token {
		p := strings.Split(t, ":=")
		amw.tokenUsers[p[1]] = p[0]
	}

	//Populate allow ip
	//amw.allowIPs["::1"] = "local"
	for _, l := range conf.Allow_ip {
		p := strings.Split(l, ":=")
		amw.allowIPs[p[1]] = p[0]
	}

}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if user, found := amw.tokenUsers[token]; found {
			// We found the token in our map
			logger.Printf("Authenticated user %s\n", user)
			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		} else if location, found := amw.allowIPs[ip]; found {
			// We found the ip in our allow ip
			logger.Printf("Authenticated from %s\n", location)
			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		} else {
			// Write an error and stop the handler chain
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
