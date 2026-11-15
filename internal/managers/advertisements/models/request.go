package models

// CreateAdvertisementRequest содержит поля запроса создания объявления
type CreateAdvertisementRequest struct {
	Title       *string
	Description *string
	Price       *float64
	Quantity    *int
	PhotoPaths  []string
}
