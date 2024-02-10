package entities

type Cart struct {
	Id     uint
	UserId uint
}
type Cart_items struct {
	Id        uint
	CartId    uint
	ProductId uint
	Quantity  int
	Total     float64
}
