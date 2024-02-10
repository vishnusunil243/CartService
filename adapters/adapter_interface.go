package adapters

import entities "github.com/vishnusunil243/CartService/entity"

type AdapterInterface interface {
	CreateCart(userId int) error
	AddToCart(req entities.Cart_items, userId int) error
	RemoveFromCart(req entities.Cart_items, userId int) error
}