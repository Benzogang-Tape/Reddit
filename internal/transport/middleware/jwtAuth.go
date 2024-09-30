package middleware

import (
	"context"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

type HTTPMethods []string
type Endpoints map[*regexp.Regexp]HTTPMethods

var (
	authUrls = Endpoints{
		regexp.MustCompile(`^/api/posts$`):                            {http.MethodPost},   // 4
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+$`):               {http.MethodPost},   // 7
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/[0-9a-fA-F-]+$`): {http.MethodDelete}, // 8
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/upvote$`):        {http.MethodGet},    // 9
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/downvote$`):      {http.MethodGet},    // 10
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/unvote$`):        {http.MethodGet},    // 11
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+$`):               {http.MethodDelete}, // 12
	}
)

func Auth(next http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var canBeWithoutAuth = true
		for endpoint, methods := range authUrls {
			if endpoint.MatchString(r.URL.Path) && slices.Contains(methods, r.Method) {
				canBeWithoutAuth = false
				break
			}
		}
		if canBeWithoutAuth {
			next.ServeHTTP(w, r)
			return
		}

		authToken := models.Session{}
		authToken.InitWithToken(strings.Split(r.Header.Get("Authorization"), " ")[1])
		payload, err := authToken.ValidateToken()
		if err != nil {
			logger.Warnw("Authorization failed",
				"reason", err.Error(),
				"remote_addr", r.RemoteAddr,
				"url", r.URL.Path,
			)
			http.Redirect(w, r, "/api/posts/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), models.Payload, payload)))
	})
}
