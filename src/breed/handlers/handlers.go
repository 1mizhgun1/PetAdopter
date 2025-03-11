package handlers

import (
	"encoding/json"
	goerrors "errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"pet_adopter/src/breed"
	"pet_adopter/src/utils"
)

type BreedHandler struct {
	logic breed.BreedLogic
}

func NewBreedHandler(logic breed.BreedLogic) *BreedHandler {
	return &BreedHandler{logic: logic}
}

func (h *BreedHandler) GetBreeds(w http.ResponseWriter, r *http.Request) {
	var (
		breeds []breed.Breed
		err    error
	)

	animalIDString := r.URL.Query().Get("animal_id")
	if animalIDString == "" {
		breeds, err = h.logic.GetBreeds(r.Context())
	} else {
		animalID, err := uuid.FromString(animalIDString)
		if err != nil {
			utils.LogError(r.Context(), err, "invalid animal id")
			http.Error(w, "invalid animal id", http.StatusBadRequest)
			return
		}

		breeds, err = h.logic.GetBreedsByAnimalID(r.Context(), animalID)
	}
	if err != nil {
		utils.LogError(r.Context(), err, "failed to get breeds")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(breeds); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func (h *BreedHandler) GetBreedByID(w http.ResponseWriter, r *http.Request) {
	breedIDString := mux.Vars(r)["id"]
	breedID, err := uuid.FromString(breedIDString)
	if err != nil {
		utils.LogError(r.Context(), err, "invalid breed id")
		http.Error(w, "invalid breed id", http.StatusBadRequest)
		return
	}

	breedData, err := h.logic.GetBreedByID(r.Context(), breedID)
	if err != nil {
		if goerrors.Is(err, breed.ErrBreedNotFound) {
			utils.LogError(r.Context(), err, "breed not found")
			http.Error(w, "breed not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get breed")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	if err = json.NewEncoder(w).Encode(breedData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type AddBreedRequest struct {
	Name     string    `json:"name"`
	AnimalID uuid.UUID `json:"animal_id"`
}

type AddBreedResponse struct {
	Breed breed.Breed `json:"breed"`
}

func (h *BreedHandler) AddBreed(w http.ResponseWriter, r *http.Request) {
	var req AddBreedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	breedData, err := h.logic.AddBreed(r.Context(), req.Name, req.AnimalID)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to add breed")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	resp := AddBreedResponse{Breed: breedData}
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type RemoveBreedRequest struct {
	ID uuid.UUID `json:"id"`
}

func (h *BreedHandler) RemoveBreedByID(w http.ResponseWriter, r *http.Request) {
	var req RemoveBreedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.logic.RemoveBreedByID(r.Context(), req.ID); err != nil {
		if goerrors.Is(err, breed.ErrBreedNotFound) {
			utils.LogError(r.Context(), err, "breed not found")
			http.Error(w, "breed not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get breed")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}
}
