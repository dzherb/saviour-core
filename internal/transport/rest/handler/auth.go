package handler

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"saviour/internal/logger"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest/api"
)

const refreshTokenCookieName = "refresh_token"

type AuthHandler struct {
	log            *slog.Logger
	authService    auth.Service
	trustedProxies []netip.Prefix
}

func NewAuthHandler(
	log *slog.Logger,
	authService auth.Service,
	trustedProxies []netip.Prefix,
) *AuthHandler {
	return &AuthHandler{
		log:            log,
		authService:    authService,
		trustedProxies: trustedProxies,
	}
}

func (h *AuthHandler) CreateSession(
	w http.ResponseWriter,
	r *http.Request,
	request api.CreateSessionRequestObject,
) (api.CreateSessionResponseObject, error) {
	tokenPair, err := h.authService.CreateSession(
		r.Context(),
		auth.CreateSessionParams{
			Username:  request.Body.Username,
			Password:  request.Body.Password,
			UserAgent: r.UserAgent(),
			IP:        clientIP(r, h.trustedProxies),
		})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, api.NewErrorResponse(
				http.StatusUnauthorized,
				api.InvalidCredentialsErrorType,
				"invalid user credentials",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling auth service on CreateSession request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	setRefreshTokenCookie(w, tokenPair.RefreshToken)

	return api.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	}, nil
}

func (h *AuthHandler) RefreshSession( //nolint:funlen
	w http.ResponseWriter,
	r *http.Request,
	_ api.RefreshSessionRequestObject,
) (api.RefreshSessionResponseObject, error) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return nil, api.NewErrorResponse(
			http.StatusUnauthorized,
			api.RefreshTokenNotSetErrorType,
			"Refresh token cookie not set",
		)
	}

	refreshToken := cookie.Value

	tokenPair, err := h.authService.RefreshSession(
		r.Context(),
		auth.RefreshSessionParams{
			RefreshToken: refreshToken,
			UserAgent:    r.UserAgent(),
			IP:           clientIP(r, h.trustedProxies),
		},
	)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return nil, api.NewErrorResponse(
				http.StatusUnauthorized,
				api.RefreshTokenExpiredErrorType,
				"Refresh token expired, try to create a new session",
			)
		}

		if errors.Is(err, auth.ErrInvalidToken) {
			h.log.WarnContext(
				r.Context(),
				"client tried to refresh session with invalid token",
				slog.String(
					"client_ip",
					clientIP(r, h.trustedProxies).String(),
				),
				slog.String("refresh_token", refreshToken),
				logger.ErrorAttr(err),
			)

			return nil, api.NewErrorResponse(
				http.StatusUnauthorized,
				api.RefreshTokenNotValidErrorType,
				"Refresh token not valid",
			)
		}

		if errors.Is(err, auth.ErrSessionNotFound) {
			return nil, api.NewErrorResponse(
				http.StatusNotFound,
				api.EntityNotFoundErrorType,
				"Associated session not found, it could be revoked "+
					"or refresh token is already used",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling auth service on RefreshSession request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	setRefreshTokenCookie(w, tokenPair.RefreshToken)

	return api.AccessTokenResponse{
		AccessToken: tokenPair.AccessToken,
	}, nil
}

func (h *AuthHandler) RevokeSession(
	_ http.ResponseWriter,
	r *http.Request,
	request api.RevokeSessionRequestObject,
) (api.RevokeSessionResponseObject, error) {
	token := auth.MustGetTokenFromContext(r.Context())

	err := h.authService.RevokeSession(
		r.Context(),
		auth.RevokeSessionParams{
			UserUUID:    token.UserUUID,
			SessionUUID: request.SessionID,
		},
	)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			return nil, api.NewErrorResponse(
				http.StatusNotFound,
				api.EntityNotFoundErrorType,
				"Session not found",
			)
		}

		h.log.ErrorContext(
			r.Context(),
			"unexpected error calling auth service on RevokeSession request",
			logger.ErrorAttr(err),
		)

		return nil, api.DefaultInternalError
	}

	return api.EmptyResponse{}, nil
}

func setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     api.RefreshSessionPath,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(auth.RefreshTokenTTL.Seconds()),
	})
}

var fallbackIP = netip.MustParseAddr("0.0.0.0")

func clientIP(r *http.Request, trustedProxies []netip.Prefix) netip.Addr {
	if r == nil {
		return fallbackIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	remoteIP, err := netip.ParseAddr(host)
	if err != nil {
		return fallbackIP
	}

	trusted := false

	for _, n := range trustedProxies {
		if n.Contains(remoteIP) {
			trusted = true
			break
		}
	}

	if !trusted {
		return remoteIP
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ipStr := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip, err := netip.ParseAddr(ipStr); err == nil {
			return ip
		}
	}

	if xrip := r.Header.Get("X-Real-Ip"); xrip != "" {
		if ip, err := netip.ParseAddr(xrip); err == nil {
			return ip
		}
	}

	return remoteIP
}
