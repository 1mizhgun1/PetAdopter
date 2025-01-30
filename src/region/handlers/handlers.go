package handlers

import (
	"encoding/json"
	goerrors "errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"pet_adopter/src/region"
	"pet_adopter/src/region/logic"
	"pet_adopter/src/utils"
)

type RegionHandler struct {
	logic logic.RegionLogic
}

func NewRegionHandler(logic logic.RegionLogic) *RegionHandler {
	return &RegionHandler{logic: logic}
}

func (h *RegionHandler) GetRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.logic.GetRegions(r.Context())
	if err != nil {
		utils.LogError(r.Context(), err, "failed to get Regions")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(regions); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func (h *RegionHandler) GetRegionByID(w http.ResponseWriter, r *http.Request) {
	regionIDString := mux.Vars(r)["id"]
	regionID, err := uuid.FromString(regionIDString)
	if err != nil {
		utils.LogError(r.Context(), err, "invalid Region id")
		http.Error(w, "invalid Region id", http.StatusBadRequest)
		return
	}

	regionData, err := h.logic.GetRegionByID(r.Context(), regionID)
	if err != nil {
		if goerrors.Is(err, region.ErrRegionNotFound) {
			utils.LogError(r.Context(), err, "Region not found")
			http.Error(w, "Region not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get Region")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	if err = json.NewEncoder(w).Encode(regionData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type AddRegionRequest struct {
	Name string `json:"name"`
}

func (h *RegionHandler) AddRegion(w http.ResponseWriter, r *http.Request) {
	var req AddRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	regionData, err := h.logic.AddRegion(r.Context(), req.Name)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to add Region")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(regionData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type RemoveRegionRequest struct {
	ID uuid.UUID `json:"id"`
}

func (h *RegionHandler) RemoveRegionByID(w http.ResponseWriter, r *http.Request) {
	var req RemoveRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.logic.RemoveRegionByID(r.Context(), req.ID); err != nil {
		if goerrors.Is(err, region.ErrRegionNotFound) {
			utils.LogError(r.Context(), err, "Region not found")
			http.Error(w, "Region not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get Region")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}
}
