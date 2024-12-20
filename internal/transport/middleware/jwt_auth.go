package middleware

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	"github.com/Benzogang-Tape/Reddit/internal/service"
)

type HTTPMethods []string
type Endpoints map[*regexp.Regexp]HTTPMethods

var (
	authUrls = Endpoints{
		regexp.MustCompile(`^/api/posts$`):                            {http.MethodPost},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+$`):               {http.MethodPost},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/[0-9a-fA-F-]+$`): {http.MethodDelete},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/upvote$`):        {http.MethodGet},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/downvote$`):      {http.MethodGet},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+/unvote$`):        {http.MethodGet},
		regexp.MustCompile(`^/api/post/[0-9a-fA-F-]+$`):               {http.MethodDelete},
	}
)

func Auth(next http.Handler, sessMngr service.SessionAPI, logger *zap.SugaredLogger) http.Handler {
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

		session := &jwt.Session{
			Token: strings.Split(r.Header.Get("Authorization"), " ")[1],
		}
		payload, err := sessMngr.Verify(r.Context(), session)
		if err != nil {
			logger.Warnw("Authorization failed",
				"reason", err.Error(),
				"remote_addr", r.RemoteAddr,
				"url", r.URL.Path,
			)

			http.Redirect(w, r, "/api/posts/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), jwt.Payload, payload)))
	})
}
