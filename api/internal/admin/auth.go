package admin

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/wolfandwings/api/internal/model"
	"github.com/wolfandwings/api/internal/store"
)

var errNoSession = errors.New("no session")

// dummyBcryptHash — для логина с несуществующим email: ответ тратит то же
// время на bcrypt, что и для существующего (защита от перечисления по таймингу).
var dummyBcryptHash, _ = bcrypt.GenerateFromPassword([]byte("wolfandwings-timing"), bcrypt.DefaultCost)

const (
	sessionCookie = "ww_admin"
	sessionTTL    = 12 * time.Hour
	loginCookie   = "ww_csrf"
)

type session struct {
	userID  int64
	csrf    string
	expires time.Time
}

type Auth struct {
	store *store.Store
	mu    sync.Mutex
	sess  map[string]*session
}

func NewAuth(s *store.Store) *Auth {
	return &Auth{store: s, sess: map[string]*session{}}
}

func (a *Auth) VerifyPassword(u *model.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (a *Auth) newSession(userID int64) (*session, string) {
	tok := randHex(24)
	csrf := randHex(16)
	s := &session{userID: userID, csrf: csrf, expires: time.Now().Add(sessionTTL)}
	a.mu.Lock()
	a.sess[tok] = s
	a.mu.Unlock()
	return s, tok
}

func (a *Auth) sessionFromReq(r *http.Request) (*session, *model.User, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil, nil, errNoSession
	}
	a.mu.Lock()
	s, ok := a.sess[c.Value]
	if ok && time.Now().After(s.expires) {
		delete(a.sess, c.Value) // просроченную сессию вычищаем сразу
		ok = false
	}
	a.mu.Unlock()
	if !ok {
		return nil, nil, errNoSession
	}
	u, err := a.store.GetUserByID(r.Context(), s.userID)
	if err != nil {
		return nil, nil, errNoSession
	}
	return s, u, nil
}

// Админка доступна только через SSH-туннель (HTTP на 127.0.0.1),
// поэтому cookie не Secure — шифрование обеспечивает SSH, не TLS.
func (a *Auth) setCookie(w http.ResponseWriter, tok string) {
	// #nosec G124 — Secure=false осознанно: админка за SSH-туннелем, TLS нет
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: tok, Path: "/",
		HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode,
		MaxAge: int(sessionTTL.Seconds()),
	})
}

func (a *Auth) clearCookie(w http.ResponseWriter) {
	// #nosec G124 — как в setCookie: админка за SSH-туннелем
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode})
}

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.mu.Lock()
		delete(a.sess, c.Value)
		a.mu.Unlock()
	}
	a.clearCookie(w)
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// loginCSRF — double-submit cookie для формы логина: сессии ещё нет,
// поэтому токен кладётся в краткоживущую куку и в скрытое поле формы.
func (a *Auth) loginCSRF(w http.ResponseWriter) string {
	tok := randHex(16)
	http.SetCookie(w, &http.Cookie{
		Name: loginCookie, Value: tok, Path: "/admin/login",
		HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode,
		MaxAge: 600,
	})
	return tok
}

func (a *Auth) checkLoginCSRF(r *http.Request) bool {
	c, err := r.Cookie(loginCookie)
	return err == nil && c.Value != "" && c.Value == r.FormValue("csrf")
}
