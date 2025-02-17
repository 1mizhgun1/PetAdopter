package logic

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
	"pet_adopter/src/utils"
)

type AdLogic struct {
	repo ad.AdRepo
}

func NewAdLogic(repo ad.AdRepo) AdLogic {
	return AdLogic{repo: repo}
}

func (l *AdLogic) SearchAds(ctx context.Context, params ad.SearchParams) ([]ad.Ad, error) {
	return l.repo.SearchAds(ctx, params)
}

func (l *AdLogic) GetAd(ctx context.Context, id uuid.UUID) (ad.Ad, error) {
	return l.repo.GetAd(ctx, id)
}

func (l *AdLogic) CreateAd(ctx context.Context, form ad.AdForm) (ad.Ad, error) {
	now := time.Now().Local().Unix()
	result := ad.Ad{
		ID:        uuid.NewV4(),
		OwnerID:   utils.GetUserIDFromContext(ctx),
		Status:    ad.Actual,
		AdForm:    form,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := l.repo.CreateAd(ctx, result); err != nil {
		return result, errors.Wrap(err, "failed to create ad")
	}

	return result, nil
}

func (l *AdLogic) UpdateAd(ctx context.Context, id uuid.UUID, form ad.UpdateForm) (ad.Ad, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.Ad{}, ad.ErrNotOwner
	}

	if err = l.repo.UpdateAd(ctx, id, form, now); err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to update ad")
	}

	result, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to get result ad")
	}

	return result, nil
}

func (l *AdLogic) UpdatePhoto(ctx context.Context, id uuid.UUID, newPhotoURL string) (ad.Ad, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.Ad{}, ad.ErrNotOwner
	}

	if err = l.repo.UpdateAd(ctx, id, ad.UpdateForm{PhotoURL: &newPhotoURL}, now); err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to update ad")
	}

	currentAd.PhotoURL = newPhotoURL
	return currentAd, nil
}

func (l *AdLogic) Close(ctx context.Context, id uuid.UUID, status ad.AdStatus) (ad.Ad, error) {
	now := time.Now().Local()

	currentAd, err := l.repo.GetAd(ctx, id)
	if err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to check owner")
	}

	if currentAd.OwnerID != utils.GetUserIDFromContext(ctx) {
		return ad.Ad{}, ad.ErrNotOwner
	}

	if err = l.repo.UpdateAd(ctx, id, ad.UpdateForm{Status: &status}, now); err != nil {
		return ad.Ad{}, errors.Wrap(err, "failed to update ad")
	}

	currentAd.Status = status
	return currentAd, nil
}

func (l *AdLogic) Delete(ctx context.Context, id uuid.UUID) error {
	return l.repo.DeleteAd(ctx, id)
}
