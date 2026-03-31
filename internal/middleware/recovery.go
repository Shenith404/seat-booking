package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/shenith404/seat-booking/internal/common"
)

// Recovery middleware recovers from panics and returns a proper error response
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v\n%s", err, debug.Stack())
				common.Err(w, common.NewInternalError("Internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
