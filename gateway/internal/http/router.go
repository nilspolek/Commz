package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/gateway/internal/prometheus"
	"team6-managing.mni.thm.de/Commz/gateway/internal/utils"
	"team6-managing.mni.thm.de/Commz/gateway/internal/ws"
)

var (
	logger = utils.GetLogger("http")
)

type Router struct {
	Router *mux.Router
}

func New(chatService, authService, aiService, mediaService, databaseUrl string) (*Router, error) {
	logger.Info().Msg("Registering routes")

	router := mux.NewRouter()
	r := &Router{
		Router: router,
	}

	chatUrl, err := url.Parse(chatService)
	if err != nil {
		return nil, err
	}

	authUrl, err := url.Parse(authService)
	if err != nil {
		return nil, err
	}

	aiUrl, err := url.Parse(aiService)
	if err != nil {
		return nil, err
	}

	mediaUrl, err := url.Parse(mediaService)
	if err != nil {
		return nil, err
	}

	// create a new hub for ws connections
	hub := ws.New()
	crawler, err := ws.NewCrawler(databaseUrl, hub)
	if err != nil {
		return nil, err
	}

	go crawler.Run()
	go hub.Run()

	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(utils.VERSION))
	})

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug().Str("url", r.URL.String()).Msg("incoming request")
			url := r.URL.String()
			defer func(start time.Time) {
				var service string = "null"
				if strings.HasPrefix(url, "/ws") {
					service = "ws"
				}
				if strings.HasPrefix(url, "/chat") {
					service = "chat"
				}
				if strings.HasPrefix(url, "/auth") {
					service = "auth"
				}
				if strings.HasPrefix(url, "/ai") {
					service = "ai"
				}
				if strings.HasPrefix(url, "/media") {
					service = "media"
				}
				prometheus.Requests.WithLabelValues(service, url).Inc()
				prometheus.ResponseTime.WithLabelValues(service, url).Observe(time.Since(start).Seconds())
			}(time.Now())
			next.ServeHTTP(w, r)
		})
	})

	router.PathPrefix("/chat").HandlerFunc(ProxyRequestHandler(chatUrl, "/chat"))
	router.PathPrefix("/auth").HandlerFunc(ProxyRequestHandler(authUrl, "/auth"))
	router.PathPrefix("/ai").HandlerFunc(ProxyRequestHandler(aiUrl, "/ai"))
	router.PathPrefix("/media").HandlerFunc(ProxyRequestHandler(mediaUrl, "/media"))

	return r, nil
}

func ProxyRequestHandler(url *url.URL, endpoint string) func(http.ResponseWriter, *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.URL.String()

		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		path := r.URL.Path
		r.URL.Path = strings.TrimPrefix(path, endpoint)

		defer func(start time.Time) {
			logger.Info().Str("from", origin).Str("to", r.URL.String()).Str("took", time.Since(start).String()).Msg("redirect")
		}(time.Now())
		proxy.ServeHTTP(w, r)
	}
}
