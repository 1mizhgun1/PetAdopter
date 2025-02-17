package handlers

import (
	"encoding/json"
	goerrors "errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"pet_adopter/src/locality"
	"pet_adopter/src/utils"
)

type LocalityHandler struct {
	logic locality.LocalityLogic
}

func NewLocalityHandler(logic locality.LocalityLogic) *LocalityHandler {
	return &LocalityHandler{logic: logic}
}

func (h *LocalityHandler) GetLocalities(w http.ResponseWriter, r *http.Request) {
	var (
		localities []locality.Locality
		err        error
	)

	regionIDString := r.URL.Query().Get("region_id")
	if regionIDString == "" {
		localities, err = h.logic.GetLocalities(r.Context())
	} else {
		regionID, err := uuid.FromString(regionIDString)
		if err != nil {
			utils.LogError(r.Context(), err, "invalid region id")
			http.Error(w, "invalid region id", http.StatusBadRequest)
			return
		}

		localities, err = h.logic.GetLocalitiesByRegionID(r.Context(), regionID)
	}
	if err != nil {
		utils.LogError(r.Context(), err, "failed to get localities")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(localities); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func (h *LocalityHandler) GetLocalityByID(w http.ResponseWriter, r *http.Request) {
	localityIDString := mux.Vars(r)["id"]
	localityID, err := uuid.FromString(localityIDString)
	if err != nil {
		utils.LogError(r.Context(), err, "invalid locality id")
		http.Error(w, "invalid locality id", http.StatusBadRequest)
		return
	}

	localityData, err := h.logic.GetLocalityByID(r.Context(), localityID)
	if err != nil {
		if goerrors.Is(err, locality.ErrLocalityNotFound) {
			utils.LogError(r.Context(), err, "locality not found")
			http.Error(w, "locality not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get locality")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	if err = json.NewEncoder(w).Encode(localityData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type AddLocalityRequest struct {
	Name      string    `json:"name"`
	RegionID  uuid.UUID `json:"region_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (h *LocalityHandler) AddLocality(w http.ResponseWriter, r *http.Request) {
	var req AddLocalityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	localityData, err := h.logic.AddLocality(r.Context(), req.Name, req.RegionID, req.Latitude, req.Longitude)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to add locality")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(localityData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type RemoveLocalityRequest struct {
	ID uuid.UUID `json:"id"`
}

func (h *LocalityHandler) RemoveLocalityByID(w http.ResponseWriter, r *http.Request) {
	var req RemoveLocalityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.logic.RemoveLocalityByID(r.Context(), req.ID); err != nil {
		if goerrors.Is(err, locality.ErrLocalityNotFound) {
			utils.LogError(r.Context(), err, "locality not found")
			http.Error(w, "locality not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get locality")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}
}
