package logic

import (
	"context"
	goerrors "errors"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/animal"
	"pet_adopter/src/breed"
	"pet_adopter/src/locality"
	"pet_adopter/src/user"
	"pet_adopter/src/utils"
)

type AdLogic struct {
	repo         ad.AdRepo
	userRepo     user.UserRepo
	animalRepo   animal.AnimalRepo
	breedRepo    breed.BreedRepo
	localityRepo locality.LocalityRepo
}

func NewAdLogic(repo ad.AdRepo, userRepo user.UserRepo, animalRepo animal.AnimalRepo, breedRepo breed.BreedRepo, localityRepo locality.LocalityRepo) AdLogic {
	return AdLogic{
		repo:         repo,
		userRepo:     userRepo,
		animalRepo:   animalRepo,
		breedRepo:    breedRepo,
		localityRepo: localityRepo,
	}
}

func (l *AdLogic) SearchAds(ctx context.Context, params ad.SearchParams) ([]ad.RespAd, error) {
	return l.repo.SearchAds(ctx, params)
}

func (l *AdLogic) GetAd(ctx context.Context, id uuid.UUID) (ad.RespAd, error) {
	return l.repo.GetAd(ctx, id)
}

func (l *AdLogic) CreateAd(ctx context.Context, form ad.AdForm, photoForm ad.PhotoParams) (ad.RespAd, error) {
	now := time.Now().Local()
	adID := uuid.NewV4()

	photoBasePath := os.Getenv("PHOTO_BASE_PATH")
	photoFilename := adID.String()

	if err := utils.WriteFileOnDisk(
		path.Join(photoBasePath, photoFilename),
		photoForm.Extension,
		photoForm.Data,
	); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to write photo on disk")
	}

	form.PhotoURL = photoFilename + photoForm.Extension

	result := ad.Ad{
		ID:        adID,
		OwnerID:   utils.GetUserIDFromContext(ctx),
		Status:    ad.Actual,
		AdForm:    form,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := l.repo.CreateAd(ctx, result); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to create ad")
	}

	return l.repo.GetAd(ctx, adID)
}

func (l *AdLogic) UpdateAd(ctx context.Context, id uuid.UUID, form ad.UpdateForm) (ad.RespAd, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.Info.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.RespAd{}, ad.ErrNotOwner
	}

	if err = l.repo.UpdateAd(ctx, id, form, now); err != nil {
		if goerrors.Is(err, ad.ErrInvalidForeignKey) {
			return ad.RespAd{}, ad.ErrInvalidForeignKey
		}
		return ad.RespAd{}, errors.Wrap(err, "failed to update ad")
	}

	return l.repo.GetAd(ctx, id)
}

func (l *AdLogic) UpdatePhoto(ctx context.Context, id uuid.UUID, photoForm ad.PhotoParams) (ad.RespAd, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.Info.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.RespAd{}, ad.ErrNotOwner
	}

	photoBasePath := os.Getenv("PHOTO_BASE_PATH")
	photoFilename := id.String()

	if err = os.Remove(path.Join(photoBasePath, currentAd.Info.PhotoURL)); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to remove old photo from disk")
	}

	if err = utils.WriteFileOnDisk(
		path.Join(photoBasePath, photoFilename),
		photoForm.Extension,
		photoForm.Data,
	); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to write new photo on disk")
	}

	newPhotoURL := photoFilename + photoForm.Extension

	if err = l.repo.UpdateAd(ctx, id, ad.UpdateForm{PhotoURL: &newPhotoURL}, now); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to update ad")
	}

	return l.repo.GetAd(ctx, id)
}

func (l *AdLogic) Close(ctx context.Context, id uuid.UUID, status string) (ad.RespAd, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.Info.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.RespAd{}, ad.ErrNotOwner
	}

	if err = l.repo.UpdateAd(ctx, id, ad.UpdateForm{Status: &status}, now); err != nil {
		return ad.RespAd{}, errors.Wrap(err, "failed to update ad")
	}

	return l.repo.GetAd(ctx, id)
}

func (l *AdLogic) Delete(ctx context.Context, id uuid.UUID) error {
	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to check owner")
	}

	photoBasePath := os.Getenv("PHOTO_BASE_PATH")

	if err = os.Remove(path.Join(photoBasePath, currentAd.Info.PhotoURL)); err != nil {
		return errors.Wrap(err, "failed to remove old photo from disk")
	}

	return l.repo.DeleteAd(ctx, id)
}
