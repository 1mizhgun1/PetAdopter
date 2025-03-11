package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	animalHandlers "pet_adopter/src/animal/handlers"
	breedHandlers "pet_adopter/src/breed/handlers"
	localityHandlers "pet_adopter/src/locality/handlers"
	regionHandlers "pet_adopter/src/region/handlers"
)

const (
	sleepDuration = 100 * time.Millisecond

	addAnimalURL   = "/api/v1/animals/add"
	addBreedURL    = "/api/v1/breeds/add"
	addRegionURL   = "/api/v1/regions/add"
	addLocalityURL = "/api/v1/localities/add"
)

var (
	adminToken = ""

	client = http.DefaultClient

	host = "http://127.0.0.1:8080"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}
	adminToken = os.Getenv("ADMIN_TOKEN")
}

type Locality struct {
	Latitude  float64
	Longitude float64
	Name      string
}

func adminRequest(path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s?token=%s", host, path, adminToken)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	if resp == nil {
		return nil, errors.New("nil response")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status=%d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()

	return respBody, nil
}

func addAnimal(name string) (uuid.UUID, error) {
	reqBody := fmt.Sprintf(`{"name":"%s"}`, name)

	respBody, err := adminRequest(addAnimalURL, strings.NewReader(reqBody))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to add animal")
	}

	resp := animalHandlers.AddAnimalResponse{}
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to unmarshal animal response")
	}

	return resp.Animal.ID, nil
}

func addBreed(name string, animalID uuid.UUID) (uuid.UUID, error) {
	reqBody := fmt.Sprintf(`{"name":"%s","animal_id":"%s"}`, name, animalID.String())

	respBody, err := adminRequest(addBreedURL, strings.NewReader(reqBody))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to add breed")
	}

	resp := breedHandlers.AddBreedResponse{}
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to unmarshal breed response")
	}

	return resp.Breed.ID, nil
}

func addRegion(name string) (uuid.UUID, error) {
	reqBody := fmt.Sprintf(`{"name":"%s"}`, name)

	respBody, err := adminRequest(addRegionURL, strings.NewReader(reqBody))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to add region")
	}

	resp := regionHandlers.AddRegionResponse{}
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to unmarshal region response")
	}

	return resp.Region.ID, nil
}

func addLocality(locality Locality, regionID uuid.UUID) (uuid.UUID, error) {
	reqBody := fmt.Sprintf(`{"name":"%s","region_id":"%s","latitude":%f,"longitude":%f}`, locality.Name, regionID, locality.Latitude, locality.Longitude)

	respBody, err := adminRequest(addLocalityURL, strings.NewReader(reqBody))
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to add locality")
	}

	resp := localityHandlers.AddLocalityResponse{}
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to unmarshal locality response")
	}

	return resp.Locality.ID, nil
}

func main() {
	flag.StringVar(&host, "host", "http://127.0.0.1:8080", "API host in format \"schema://host:port\"")
	flag.Parse()

	//for animal, breedSlice := range breeds {
	//	animalID, err := addAnimal(animal)
	//	if err != nil {
	//		log.Printf("failed to add animal: %v\n", err)
	//		continue
	//	}
	//
	//	time.Sleep(sleepDuration)
	//
	//	for _, breed := range breedSlice {
	//		_, err = addBreed(breed, animalID)
	//		if err != nil {
	//			log.Printf("failed to add breed: %v\n", err)
	//		}
	//
	//		time.Sleep(sleepDuration)
	//	}
	//}

	for region, localitySlice := range localities {
		regionID, err := addRegion(region)
		if err != nil {
			log.Printf("failed to add region: %v\n", err)
			continue
		}

		time.Sleep(sleepDuration)

		for _, locality := range localitySlice {
			_, err = addLocality(locality, regionID)
			if err != nil {
				log.Printf("failed to add locality: %v\n", err)
			}

			time.Sleep(sleepDuration)
		}
	}
}
