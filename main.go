package main

import (
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
)

const defaultPort = "8080"

type OAuthResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Team  struct {
		ID string `json:"id"`
	} `json:"team"`
	User *User `json:"user"`
}

type Server struct {
	sm *scs.SessionManager

	allowedTeams map[string]struct{}
	clientID     string
	clientSecret string
	redirectURL  string
}

type User struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Email string `json:"email"`
}

func init() {
	gob.Register(User{})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	sm := scs.New()
	sm.Cookie.Name = "session"
	sm.Cookie.Persist = true
	sm.IdleTimeout = 30 * time.Minute
	sm.Lifetime = 3 * time.Hour
	//sm.Cookie.Domain = "example.com"
	//sm.Cookie.HttpOnly = true
	//sm.Cookie.SameSite = http.SameSiteStrictMode
	//sm.Cookie.Secure = true

	s := &Server{
		sm:           sm,
		allowedTeams: map[string]struct{}{},
		clientID:     os.Getenv("SLACKBOT_OAUTH_CLIENT_ID"),
		clientSecret: os.Getenv("SLACKBOT_OAUTH_CLIENT_SECRET"),
		redirectURL:  "http://localhost:8080/oauth/slack",
	}
	allowedTeams := os.Getenv("SLACKBOT_ALLOWED_TEAM_IDS")
	for _, id := range strings.Split(allowedTeams, ",") {
		s.allowedTeams[id] = struct{}{}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/oauth/slack", s.oauthHandler)
	handler := sm.LoadAndSave(mux)

	log.Printf("connect to http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, handler))
}

var login = template.Must(template.New("login").Parse(`<html>
<a href="{{.url}}"><img src="https://api.slack.com/img/sign_in_with_slack.png" /></a>
`))

var logout = template.Must(template.New("logout").Parse(`<html>
User: {{.ID}}<br />
&nbsp;&nbsp;{{.Name}} &lt;{{.Email}}&gt;<br />
<br />
<a href="/logout">Log Out</a>
`))

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := s.sm.Get(r.Context(), "user").(User)
	if ok && user.ID != "" {
		err := logout.Execute(w, user)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		q := url.Values{}
		q.Set("scope", "identity.basic,identity.email")
		q.Set("client_id", s.clientID)
		q.Set("redirect_url", s.redirectURL)
		url := url.URL{
			Scheme:   "https",
			Host:     "slack.com",
			Path:     "/oauth/authorize",
			RawQuery: q.Encode(),
		}

		err := login.Execute(w, map[string]string{
			"url": url.String(),
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := s.sm.Clear(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = s.sm.RenewToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func (s *Server) oauthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO handle error=
	q := r.URL.Query()
	code := q.Get("code")

	q = url.Values{}
	q.Set("client_id", s.clientID)
	q.Set("client_secret", s.clientSecret)
	q.Set("code", code)
	url := url.URL{
		Scheme:   "https",
		Host:     "slack.com",
		Path:     "/api/oauth.access",
		RawQuery: q.Encode(),
	}

	res, err := http.Get(url.String())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmp := &OAuthResponse{}
	json.Unmarshal(raw, tmp)
	// TODO handle tmp.Error

	if tmp.Ok {
		if len(s.allowedTeams) > 0 {
			if _, ok := s.allowedTeams[tmp.Team.ID]; !ok {
				http.Error(w, "access denied", 403)
				return
			}
		}
		err := s.sm.RenewToken(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.sm.Put(r.Context(), "user", tmp.User)
	}

	http.Redirect(w, r, "/", 302)
}
