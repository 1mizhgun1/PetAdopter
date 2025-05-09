package handlers

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/chatgpt"
	"pet_adopter/src/config"
	"pet_adopter/src/locality"
	"pet_adopter/src/user"
	"pet_adopter/src/utils"
)

type AdHandler struct {
	logic         ad.AdLogic
	userLogic     user.UserLogic
	localityLogic locality.LocalityLogic
	chatGPT       chatgpt.ChatGPT
	cfg           config.AdConfig
}

func NewAdHandler(logic ad.AdLogic, userLogic user.UserLogic, localityLogic locality.LocalityLogic, chatGPT chatgpt.ChatGPT, cfg config.AdConfig) *AdHandler {
	return &AdHandler{
		logic:         logic,
		userLogic:     userLogic,
		localityLogic: localityLogic,
		chatGPT:       chatGPT,
		cfg:           cfg,
	}
}

type SearchResponse struct {
	Ads []ad.RespAd `json:"ads"`
}

func (h *AdHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	searchParams, err := getSearchParamsFromQuery(r.URL.Query(), h.cfg)
	if err != nil {
		utils.LogError(ctx, err, "failed to parse search params")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	searchExtra := ad.SearchExtra{}
	if searchParams.Radius != nil {
		userID := utils.GetUserIDFromContext(ctx)
		userInfo, err := h.userLogic.GetUserByID(ctx, userID)
		if err == nil {
			loc, err := h.localityLogic.GetLocalityByID(ctx, userInfo.LocalityID)
			if err == nil {
				searchExtra.Latitude = loc.Latitude
				searchExtra.Longitude = loc.Longitude
			} else {
				utils.LogError(ctx, err, "failed to get user location")
				searchParams.Radius = nil
			}
		} else {
			utils.LogError(ctx, err, fmt.Sprintf("failed to get user by ID=%s", userID.String()))
			searchParams.Radius = nil
		}
	}

	foundAds, err := h.logic.SearchAds(ctx, searchParams, searchExtra)
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
	Ad ad.RespAd `json:"ad"`
}

func (h *AdHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	foundAd, err := h.logic.GetAd(ctx, adID)
	if err != nil {
		handleAdError(ctx, w, err)
		return
	}

	result := GetResponse{Ad: foundAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type GetSameResponse struct {
	Ads []ad.RespAd `json:"ads"`
}

func (h *AdHandler) GetSame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	desc, err := h.chatGPT.GetDescriptionFromDB(ctx, adID)
	if err != nil {
		if goerrors.Is(err, chatgpt.ErrDescriptionNotFound) {
			utils.LogError(ctx, err, "description not found")
			http.Error(w, utils.NotFound, http.StatusNotFound)
		} else {
			utils.LogError(ctx, err, "failed to get description")
			http.Error(w, utils.Internal, http.StatusInternalServerError)
		}
		return
	}

	color, err := utils.ParseColor(desc.Color)
	if err != nil {
		utils.LogError(ctx, err, "failed to parse color")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	same, err := h.chatGPT.GetSame(ctx, adID, color)
	if err != nil {
		utils.LogError(ctx, err, "failed to get same ads")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	result := GetSameResponse{Ads: same}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type CreateResponse struct {
	Ad ad.RespAd `json:"ad"`
}

func (h *AdHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	photoData := h.getPhotoDataFromRequest(w, r)
	if photoData == nil {
		return
	}

	adFormJSON := []byte(r.FormValue(h.cfg.CreateFormFieldName))
	var adForm ad.AdForm
	if err := json.Unmarshal(adFormJSON, &adForm); err != nil {
		utils.LogError(ctx, err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	createdAd, err := h.logic.CreateAd(ctx, adForm, *photoData)
	if err != nil {
		handleAdError(ctx, w, err)
		return
	}

	go func() {
		newCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err = h.chatGPT.DescribePhoto(newCtx, createdAd.Info.ID, *photoData, false); err != nil {
			utils.LogError(ctx, err, "failed to describe photo")
		}
	}()

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
	Ad ad.RespAd `json:"ad"`
}

func (h *AdHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
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
		handleAdError(ctx, w, err)
		return
	}

	result := UpdateResponse{Ad: updatedAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type UpdatePhotoResponse struct {
	Ad ad.RespAd `json:"ad"`
}

func (h *AdHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	photoData := h.getPhotoDataFromRequest(w, r)
	if photoData == nil {
		return
	}

	updatedAd, err := h.logic.UpdatePhoto(ctx, adID, *photoData)
	if err != nil {
		handleAdError(ctx, w, err)
		return
	}

	go func() {
		newCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err = h.chatGPT.DescribePhoto(newCtx, updatedAd.Info.ID, *photoData, true); err != nil {
			utils.LogError(ctx, err, "failed to describe photo")
		}
	}()

	result := UpdatePhotoResponse{Ad: updatedAd}
	if err = json.NewEncoder(w).Encode(result); err != nil {
		utils.LogError(ctx, err, utils.MsgErrMarshalResponse)
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

type CloseRequest struct {
	Status string `json:"status"`
}

type CloseResponse struct {
	Ad ad.RespAd `json:"ad"`
}

func (h *AdHandler) Close(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	adID, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		utils.LogError(ctx, err, "invalid ad id")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	var req CloseRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError(r.Context(), err, utils.MsgErrUnmarshalRequest)
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if req.Status != ad.Realised && req.Status != ad.Cancelled {
		utils.LogErrorMessage(r.Context(), fmt.Sprintf("invalid status: %s", string(req.Status)))
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	updateForm := ad.UpdateForm{Status: &req.Status}
	updatedAd, err := h.logic.UpdateAd(ctx, adID, updateForm)
	if err != nil {
		handleAdError(ctx, w, err)
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
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return
	}

	if err = h.logic.Delete(ctx, adID); err != nil {
		utils.LogError(ctx, err, "failed to delete ad")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}

	if err = h.chatGPT.DeleteDescription(ctx, adID); err != nil {
		utils.LogError(ctx, err, "failed to delete description")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
		return
	}
}

func getSearchParamsFromQuery(query url.Values, cfg config.AdConfig) (ad.SearchParams, error) {
	result := ad.NewSearchParams(cfg)

	allStatusesString := query.Get("all_statuses")
	if allStatusesString != "" {
		allStatuses, err := strconv.ParseBool(allStatusesString)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse all_statuses")
		}
		result.AllStatuses = allStatuses
	}

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

	radiusString := query.Get("radius")
	if radiusString == "" {
		result.Radius = nil
	} else {
		radius64, err := strconv.ParseInt(radiusString, 10, 64)
		if err != nil {
			return result, errors.Wrap(err, "failed to parse radius")
		}
		radius := int(radius64)
		result.Radius = &radius
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

func (h *AdHandler) getPhotoDataFromRequest(w http.ResponseWriter, r *http.Request) *ad.PhotoParams {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.AdPhotoConfig.MaxFormDataSize)
	defer r.Body.Close()

	if err := r.ParseMultipartForm(h.cfg.AdPhotoConfig.MaxFormDataSize); err != nil {
		utils.LogError(ctx, err, "failed to parse multipart form, too large")
		http.Error(w, utils.Invalid, http.StatusRequestEntityTooLarge)
		return nil
	}
	defer func() {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			utils.LogError(ctx, err, "failed to remove photo from multipart form")
		}
	}()

	files := r.MultipartForm.File[h.cfg.AdPhotoConfig.RequestFieldName]
	if len(files) > 1 {
		utils.LogError(ctx, goerrors.New("multipart form contains multiple files"), "failed to add multiple files")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return nil
	}

	photoFile, _, err := r.FormFile(h.cfg.AdPhotoConfig.RequestFieldName)
	if err != nil {
		utils.LogError(ctx, err, "failed to get photo file")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return nil
	}

	content, err := io.ReadAll(photoFile)
	if err != nil && !goerrors.Is(err, io.EOF) {
		if goerrors.As(err, new(*http.MaxBytesError)) {
			utils.LogError(ctx, err, "failed to read file content, too large")
			http.Error(w, utils.Invalid, http.StatusRequestEntityTooLarge)
			return nil
		}
	}

	photoFileExtension := utils.GetFormat(h.cfg.AdPhotoConfig.FileTypes, content)
	if photoFileExtension == "" {
		utils.LogError(ctx, goerrors.New("unknown file extension"), "failed to get file format")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
		return nil
	}

	return &ad.PhotoParams{
		Data:      photoFile,
		Extension: photoFileExtension,
	}
}

func handleAdError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case goerrors.Is(err, ad.ErrAdNotFound):
		utils.LogError(ctx, err, "ad not found")
		http.Error(w, utils.NotFound, http.StatusNotFound)
	case goerrors.Is(err, ad.ErrNotOwner):
		utils.LogError(ctx, err, "not owner")
		http.Error(w, utils.NotFound, http.StatusForbidden)
	case goerrors.Is(err, ad.ErrInvalidForeignKey):
		utils.LogError(ctx, err, "invalid foreign key")
		http.Error(w, utils.Invalid, http.StatusBadRequest)
	default:
		utils.LogError(ctx, err, "failed to perform operation")
		http.Error(w, utils.Internal, http.StatusInternalServerError)
	}
}
