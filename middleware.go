package main

import (
	"net"
	"net/http"
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

// Initialize it somewhere
func (amw *authenticationMiddleware) Populate() {
	amw.tokenUsers = make(map[string]string)
	amw.allowIPs = make(map[string]string)

	//Populate token
	amw.tokenUsers["00000000"] = "user0"
	amw.tokenUsers["aaaaaaaa"] = "userA"
	amw.tokenUsers["05f717e5"] = "randomUser"
	amw.tokenUsers["deadbeef"] = "user0"

	//Populate allow ip
	amw.allowIPs["::1"] = "local"
	amw.allowIPs["1.2.3.4"] = "office"

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
