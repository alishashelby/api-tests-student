package advertisements

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/managers/advertisements/models"
)

func GetAdvertisementsList(t *testing.T, token, text string) models.AdvertisementsList {
	rawResponse := advertisementsClient.HttpGetAdvertisements(t, token, text)
	defer rawResponse.Body.Close()

	body, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	response := models.AdvertisementsList{}
	require.NoError(t, json.Unmarshal(body, &response))

	return response
}
