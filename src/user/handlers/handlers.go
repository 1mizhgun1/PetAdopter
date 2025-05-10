package handlers

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/satori/uuid"
	"pet_adopter/src/config"
	"pet_adopter/src/locality"
	"pet_adopter/src/user"
	"pet_adopter/src/utils"
)

type UserHandler struct {
	user          user.UserLogic
	session       user.SessionLogic
	locality      locality.LocalityLogic
	sessionCfg    config.SessionConfig
	validationCfg config.ValidationConfig
}

func NewUserHandler(user user.UserLogic, session user.SessionLogic, locality locality.LocalityLogic, sessionCfg config.SessionConfig, validationCfg config.ValidationConfig) *UserHandler {
	return &UserHandler{
		user:          user,
		session:       session,
		locality:      locality,
		sessionCfg:    sessionCfg,
		validationCfg: validationCfg,
	}
}

func (h *UserHandler) validateUserCredentials(username string, password string) error {
	if err := utils.ValidateUsername(username, h.validationCfg); err != nil {
		return err
	}

	if err := utils.ValidatePassword(password, h.validationCfg); err != nil {
		return err
	}

	return nil
}

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	User         user.User `json:"user"`
	RefreshToken string    `json:"refresh_token"`
	Locality     string    `json:"locality"`
}

// SignUp
// @Summary	Sign up
// @Description	Add a new user to the database
// @Tags user
// @ID sign-up
// @Accept json
// @Produce	json
// @Param credentials body SignUpRequest true "request"
// @Success	200	{object} SignUpResponse	"response 200"
// @Failure	400	{object} string "response 400" "invalid"
// @Failure	500	{object} string "response 500" "internal"
// @Router /user/signup [post]
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	req := SignUpRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.validateUserCredentials(req.Username, req.Password); err != nil {
		utils.LogError(r.Context(), err, "invalid credentials")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	userData, err := h.user.CreateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		if goerrors.Is(err, user.ErrUserAlreadyExists) {
			utils.LogErrorMessage(r.Context(), user.ErrUserAlreadyExists.Error())
			http.Error(w, "user already exists", http.StatusConflict)
		} else {
			utils.LogError(r.Context(), err, "failed to create user")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	accessToken, refreshToken, err := h.session.SetSession(r.Context(), userData.Username)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to set session")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCfg.AccessTokenCookieName,
		Secure:   h.sessionCfg.ProtectedCookies,
		Value:    accessToken,
		HttpOnly: true,
		Expires:  time.Now().Local().Add(h.sessionCfg.AccessTokenLifeTime),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp := SignUpResponse{
		User:         userData,
		RefreshToken: refreshToken,
		Locality:     "",
	}
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         user.User `json:"user"`
	RefreshToken string    `json:"refresh_token"`
	Locality     string    `json:"locality"`
}

// Login
// @Summary	Login
// @Description	login
// @Tags user
// @ID login
// @Accept json
// @Produce	json
// @Param credentials body LoginRequest true "request"
// @Success	200	{object} LoginResponse "response 200"
// @Failure	400	{object} string "response 400" "invalid"
// @Failure	500	{object} string "response 500" "internal"
// @Router /user/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(ctx, err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.validateUserCredentials(req.Username, req.Password); err != nil {
		utils.LogError(ctx, err, "invalid credentials")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userData, correctPassword, err := h.user.CheckPassword(ctx, req.Username, req.Password)
	if err != nil {
		if goerrors.Is(err, user.ErrUserNotFound) {
			utils.LogErrorMessage(ctx, user.ErrUserNotFound.Error())
			http.Error(w, "incorrect username or password", http.StatusBadRequest)
		} else {
			utils.LogError(ctx, err, "failed to check password")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}
	if !correctPassword {
		utils.LogErrorMessage(ctx, "incorrect password")
		http.Error(w, "incorrect username or password", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.session.SetSession(ctx, userData.Username)
	if err != nil {
		utils.LogError(ctx, err, "failed to set session")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCfg.AccessTokenCookieName,
		Secure:   h.sessionCfg.ProtectedCookies,
		Value:    accessToken,
		HttpOnly: true,
		Expires:  time.Now().Local().Add(h.sessionCfg.AccessTokenLifeTime),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	resp := LoginResponse{
		User:         userData,
		RefreshToken: refreshToken,
	}

	if userData.LocalityID != uuid.Nil {
		loc, err := h.locality.GetLocalityByID(ctx, userData.LocalityID)
		if err != nil {
			utils.LogError(ctx, err, "failed to get locality")
			resp.Locality = ""
		} else {
			resp.Locality = loc.Name
		}
	} else {
		resp.Locality = ""
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

// Logout
// @Summary	Logout
// @Description	logout
// @Tags user
// @ID logout
// @Accept json
// @Produce	json
// @Success	200
// @Failure	401
// @Router /user/logout [post]
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Del("Authorization")
	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCfg.AccessTokenCookieName,
		Secure:   h.sessionCfg.ProtectedCookies,
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

type GetUserResponse struct {
	User user.User `json:"user"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		utils.LogErrorMessage(r.Context(), "user not found in context")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	userData, err := h.user.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to get user by id")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	resp := GetUserResponse{User: userData}
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type SetLocalityRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type SetLocalityResponse struct {
	User     user.User `json:"user"`
	Locality string    `json:"locality"`
}

func (h *UserHandler) SetLocality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := utils.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		utils.LogErrorMessage(ctx, "user not found in context")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	req := SetLocalityRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(ctx, err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	loc, err := h.locality.GetLocalityByCoords(ctx, req.Latitude, req.Longitude)
	if err != nil {
		utils.LogError(ctx, err, "failed to get locality by coords")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	userData, err := h.user.SetLocalityID(ctx, userID, loc.ID)
	if err != nil {
		if goerrors.Is(err, locality.ErrLocalityNotFound) {
			utils.LogError(ctx, err, "locality not found")
			http.Error(w, utils.Invalid, http.StatusBadRequest)
			return
		}
		utils.LogError(ctx, err, "failed to set locality")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	resp := SetLocalityResponse{User: userData, Locality: loc.Name}
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}
