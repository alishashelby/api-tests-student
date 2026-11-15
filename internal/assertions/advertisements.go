package assertions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/managers/advertisements/models"
)

func ContainsAdvertisementByTitle(t *testing.T, items []models.Advertisement, expectedTitle string) {
	t.Helper()

	for _, item := range items {
		if item.Title == expectedTitle {
			return
		}
	}

	require.Failf(t, "Объявление не найдено", "объявление с названием %q отсутствует в ответе", expectedTitle)
}

// FindAdvertisementByID ищет объявление по идентификатору
func FindAdvertisementByID(items []models.Advertisement, expectedID string) (models.Advertisement, bool) {
	for _, item := range items {
		if item.ID == expectedID {
			return item, true
		}
	}

	return models.Advertisement{}, false
}
