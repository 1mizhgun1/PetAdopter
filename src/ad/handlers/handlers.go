package handlers

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/config"
	"pet_adopter/src/utils"
)

type AdHandler struct {
	logic ad.AdLogic
	cfg   config.AdConfig
}

func NewAdHandler(logic ad.AdLogic, cfg config.AdConfig) *AdHandler {
	return &AdHandler{logic: logic, cfg: cfg}
}

type SearchResponse struct {
	Ads []ad.Ad `json:"ads"`
}

func (h *AdHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	searchParams, err := getSearchParamsFromQuery(r.URL.Query(), h.cfg)
	if err != nil {
		utils.LogError(ctx, err, "failed to parse search params")
		http.Error(w, "invalid search params", http.StatusBadRequest)
		return
	}

	foundAds, err := h.logic.SearchAds(ctx, searchParams)
	if err != nil {
		utils.LogError(ctx, err, "failed to search ads")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	result := SearchResponse{Ads: foundAds}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type GetResponse struct {
	Ad ad.Ad `json:"ad"`
}

func (h *AdHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, "invalid ad id", http.StatusBadRequest)
		return
	}

	foundAd, err := h.logic.GetAd(ctx, adID)
	if err != nil {
		if goerrors.Is(err, ad.ErrAdNotFound) {
			utils.LogError(ctx, err, "ad not found")
			http.Error(w, "ad not found", http.StatusNotFound)
		} else {
			utils.LogError(ctx, err, "failed to get ad")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	result := GetResponse{Ad: foundAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type CreateRequest struct {
	ad.AdForm
}

type CreateResponse struct {
	Ad ad.Ad `json:"ad"`
}

func (h *AdHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(ctx, err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	createdAd, err := h.logic.CreateAd(ctx, req.AdForm)
	if err != nil {
		utils.LogError(ctx, err, "failed to create ad")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	result := CreateResponse{Ad: createdAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type UpdateRequest struct {
	Form           ad.AdForm `json:"form"`
	FieldsToUpdate []string  `json:"fields_to_update"`
}

type UpdateResponse struct {
	Ad ad.Ad `json:"ad"`
}

func (h *AdHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, "invalid ad id", http.StatusBadRequest)
		return
	}

	var req UpdateRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	updateForm := getUpdateFormFromRequest(req)
	updatedAd, err := h.logic.UpdateAd(ctx, adID, updateForm)
	if err != nil {
		if goerrors.Is(err, ad.ErrAdNotFound) {
			utils.LogError(ctx, err, "ad not found")
			http.Error(w, "ad not found", http.StatusNotFound)
		} else {
			utils.LogError(ctx, err, "failed to update ad")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	result := UpdateResponse{Ad: updatedAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type CloseRequest struct {
	Status ad.AdStatus `json:"status"`
}

type CloseResponse struct {
	Ad ad.Ad `json:"ad"`
}

func (h *AdHandler) Close(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, "invalid ad id", http.StatusBadRequest)
		return
	}

	var req CloseRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	updateForm := ad.UpdateForm{Status: &req.Status}
	updatedAd, err := h.logic.UpdateAd(ctx, adID, updateForm)
	if err != nil {
		if goerrors.Is(err, ad.ErrAdNotFound) {
			utils.LogError(ctx, err, "ad not found")
			http.Error(w, "ad not found", http.StatusNotFound)
		} else {
			utils.LogError(ctx, err, "failed to close ad")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	result := CloseResponse{Ad: updatedAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func (h *AdHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, "invalid ad id", http.StatusBadRequest)
		return
	}

	if err = h.logic.Delete(ctx, adID); err != nil {
		utils.LogError(ctx, err, "failed to delete ad")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func getSearchParamsFromQuery(query url.Values, cfg config.AdConfig) (ad.SearchParams, error) {
	result := ad.SearchParams{}

	animalIDString := query.Get("animal_id")
	if animalIDString == "" {
		result.AnimalID = nil
	} else {
		animalID, err := uuid.FromString(animalIDString)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse animal_id")
		}
		result.AnimalID = &animalID
	}

	breedIDString := query.Get("breed_id")
	if breedIDString == "" {
		result.BreedID = nil
	} else {
		breedID, err := uuid.FromString(breedIDString)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse breed_id")
		}
		result.BreedID = &breedID
	}

	minPriceString := query.Get("min_price")
	if minPriceString == "" {
		result.MinPrice = nil
	} else {
		minPrice64, err := strconv.ParseInt(minPriceString, 10, 64)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse min_price")
		}
		minPrice := int(minPrice64)
		result.MinPrice = &minPrice
	}

	maxPriceString := query.Get("max_price")
	if maxPriceString == "" {
		result.MaxPrice = nil
	} else {
		maxPrice64, err := strconv.ParseInt(maxPriceString, 10, 64)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse max_price")
		}
		maxPrice := int(maxPrice64)
		result.MaxPrice = &maxPrice
	}

	limitString := query.Get("limit")
	if limitString == "" {
		result.Limit = cfg.DefaultSearchLimit
	} else {
		limit64, err := strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse limit")
		}
		limit := int(limit64)
		if limit > cfg.MaxSearchLimit {
			result.Limit = cfg.MaxSearchLimit
		}
		result.Limit = limit
	}

	offsetString := query.Get("offset")
	if offsetString == "" {
		result.Offset = cfg.DefaultSearchOffset
	} else {
		offset64, err := strconv.ParseInt(offsetString, 10, 64)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse offset")
		}
		result.Offset = int(offset64)
	}

	return result, nil
}

func getUpdateFormFromRequest(req UpdateRequest) ad.UpdateForm {
	result := ad.UpdateForm{}

	fieldsToUpdateMap := make(map[string]bool)
	for _, field := range req.FieldsToUpdate {
		fieldsToUpdateMap[field] = true
	}

	if fieldsToUpdateMap["photo_url"] {
		result.PhotoURL = &req.Form.PhotoURL
	}
	if fieldsToUpdateMap["title"] {
		result.Title = &req.Form.Title
	}
	if fieldsToUpdateMap["description"] {
		result.Description = &req.Form.Description
	}
	if fieldsToUpdateMap["animal_id"] {
		result.AnimalID = &req.Form.AnimalID
	}
	if fieldsToUpdateMap["breed_id"] {
		result.BreedID = &req.Form.BreedID
	}
	if fieldsToUpdateMap["price"] {
		result.Price = &req.Form.Price
	}
	if fieldsToUpdateMap["contacts"] {
		result.Contacts = &req.Form.Contacts
	}

	return result
}
