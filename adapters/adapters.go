package adapters

import (
	"fmt"

	entities "github.com/vishnusunil243/CartService/entity"
	"gorm.io/gorm"
)

type CartAdapter struct {
	DB *gorm.DB
}

func NewCartAdapter(db *gorm.DB) *CartAdapter {
	return &CartAdapter{DB: db}
}
func (cart *CartAdapter) CreateCart(userId int) error {
	query := `INSERT INTO carts (user_id) VALUES ($1)`
	if err := cart.DB.Exec(query, userId).Error; err != nil {
		return err
	}
	return nil

}
func (cart *CartAdapter) AddToCart(req entities.Cart_items, userId int) error {
	tx := cart.DB.Begin()
	var cartId int
	var current entities.Cart_items
	queryId := `SELECT id FROM carts WHERE user_id=?`
	if err := tx.Raw(queryId, userId).Scan(&cartId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("cart not found")
	}
	queryCurrent := `SELECT * FROM cart_items WHERE cart_id=$1 AND product_id=$2`
	if err := tx.Raw(queryCurrent, cartId, req.ProductId).Scan(&current).Error; err != nil {
		tx.Rollback()
		return err
	}
	var res entities.Cart_items
	if current.ProductId == 0 {
		insertQuery := `INSERT INTO cart_items (cart_id,product_id,quantity,total)VALUES($1,$2,$3,0) RETURNING id,product_id,cart_id`
		if err := tx.Raw(insertQuery, cartId, req.ProductId, req.Quantity).Scan(&res).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		updateQuery := `UPDATE cart_items SET quantity=quantity+$1 WHERE cart_id=$2 AND product_id=$3`
		if err := tx.Exec(updateQuery, req.Quantity, cartId, req.ProductId).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	updateTotal := `UPDATE cart_items SET total=total+$1 WHERE cart_id=$2 AND product_id=$3`
	if err := tx.Exec(updateTotal, (req.Total * float64(req.Quantity)), cartId, req.ProductId).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func (cart *CartAdapter) RemoveFromCart(req entities.Cart_items, userId int) error {
	tx := cart.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var current entities.Cart_items
	var cartId int
	queryId := `SELECT id FROM carts WHERE user_id=?`
	if err := tx.Raw(queryId, userId).Scan(&cartId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("cart not found: %v", err)
	}

	query := `SELECT * FROM cart_items WHERE cart_id=? AND product_id=?`
	if err := tx.Raw(query, cartId, req.ProductId).Scan(&current).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch cart item: %v", err)
	}

	if current.ProductId == 0 {
		tx.Rollback()
		return fmt.Errorf("this product is not present in cart")
	}

	if current.Quantity <= 0 {
		tx.Rollback()
		return fmt.Errorf("given product is not found in cart")
	}

	updateQuery := `UPDATE cart_items SET quantity=quantity-1 WHERE cart_id=? AND product_id=?`
	if err := tx.Exec(updateQuery, cartId, req.ProductId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update quantity: %v", err)
	}
	var quantity int
	updateTotal := `UPDATE cart_items SET total=total-? WHERE cart_id=? AND product_id=? RETURNING quantity`
	if err := tx.Raw(updateTotal, req.Total, cartId, req.ProductId).Scan(&quantity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update total: %v", err)
	}
	if quantity == 0 {
		deleteFromCartItems := `DELETE FROM cart_items WHERE cart_id=? AND product_id=?`
		if err := tx.Exec(deleteFromCartItems, cartId, req.ProductId).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error in deleting cart_items")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
func (cart *CartAdapter) IsEmpty(req entities.Cart_items, userId int) bool {
	query := `SELECT id FROM carts WHERE user_id=?`
	var cartId int
	if err := cart.DB.Raw(query, userId).Scan(&cartId).Error; err != nil {
		return true
	}
	emptyCheck := `SELECT * FROM cart_items WHERE cart_id=?`
	var current entities.Cart_items
	if err := cart.DB.Raw(emptyCheck, cartId).Scan(&current).Error; err != nil {
		return true
	}
	if current.CartId == 0 {
		return true
	}
	return false
}
func (cart *CartAdapter) GetAllCartItems(userId int) ([]entities.Cart_items, error) {
	var cartId int
	query := `SELECT id FROM carts WHERE user_id=?`
	if err := cart.DB.Raw(query, userId).Scan(&cartId).Error; err != nil {
		return []entities.Cart_items{}, err
	}
	var cartItems []entities.Cart_items
	if err := cart.DB.Raw(`SELECT * FROM cart_items WHERE cart_id=?`, cartId).Scan(&cartItems).Error; err != nil {
		return []entities.Cart_items{}, err
	}
	return cartItems, nil
}
