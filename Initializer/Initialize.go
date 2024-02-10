package initializer

import (
	"github.com/vishnusunil243/CartService/adapters"
	"github.com/vishnusunil243/CartService/service"
	"gorm.io/gorm"
)

func Initializer(db *gorm.DB) *service.CartService {
	adapter := adapters.NewCartAdapter(db)
	service := service.NewCartService(adapter)
	return service
}
