package handlers

import (
	"encoding/json"
	goerrors "errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"pet_adopter/src/animal"
	"pet_adopter/src/animal/logic"
	"pet_adopter/src/utils"
)

type AnimalHandler struct {
	logic logic.AnimalLogic
}

func NewAnimalHandler(logic logic.AnimalLogic) *AnimalHandler {
	return &AnimalHandler{logic: logic}
}

func (h *AnimalHandler) GetAnimals(w http.ResponseWriter, r *http.Request) {
	animals, err := h.logic.GetAnimals(r.Context())
	if err != nil {
		utils.LogError(r.Context(), err, "failed to get animals")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(animals); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func (h *AnimalHandler) GetAnimalByID(w http.ResponseWriter, r *http.Request) {
	animalIDString := mux.Vars(r)["id"]
	animalID, err := uuid.FromString(animalIDString)
	if err != nil {
		utils.LogError(r.Context(), err, "invalid animal id")
		http.Error(w, "invalid animal id", http.StatusBadRequest)
		return
	}

	animalData, err := h.logic.GetAnimalByID(r.Context(), animalID)
	if err != nil {
		if goerrors.Is(err, animal.ErrAnimalNotFound) {
			utils.LogError(r.Context(), err, "animal not found")
			http.Error(w, "animal not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get animal")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	if err = json.NewEncoder(w).Encode(animalData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type AddAnimalRequest struct {
	Name string `json:"name"`
}

func (h *AnimalHandler) AddAnimal(w http.ResponseWriter, r *http.Request) {
	var req AddAnimalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	animalData, err := h.logic.AddAnimal(r.Context(), req.Name)
	if err != nil {
		utils.LogError(r.Context(), err, "failed to add animal")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(animalData); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type RemoveAnimalRequest struct {
	ID uuid.UUID `json:"id"`
}

func (h *AnimalHandler) RemoveAnimalByID(w http.ResponseWriter, r *http.Request) {
	var req RemoveAnimalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err := h.logic.RemoveAnimalByID(r.Context(), req.ID); err != nil {
		if goerrors.Is(err, animal.ErrAnimalNotFound) {
			utils.LogError(r.Context(), err, "animal not found")
			http.Error(w, "animal not found", http.StatusNotFound)
		} else {
			utils.LogError(r.Context(), err, "failed to get animal")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}
}
