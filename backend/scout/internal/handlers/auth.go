package handlers

import (
	"net/http"
	"time"

	"scout/internal/auth"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "scout_session"

type AuthHandler struct {
	adminPassword string
	sessionSecret string
	cookieSecure  bool
}

func NewAuthHandler(adminPassword, sessionSecret string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{adminPassword: adminPassword, sessionSecret: sessionSecret, cookieSecure: cookieSecure}
}

type loginRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	if req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token := auth.NewSessionToken(h.sessionSecret, time.Now())
	c.SetCookie(sessionCookieName, token, int(auth.SessionDuration.Seconds()), "/", "", h.cookieSecure, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RequireAuth guards every route it is applied to — Scout has no public
// application-data route group (NF1), unlike oncarinho.
func RequireAuth(sessionSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(sessionCookieName)
		if err != nil || !auth.ValidateSessionToken(sessionSecret, cookie, time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
