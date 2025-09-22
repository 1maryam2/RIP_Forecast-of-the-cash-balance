package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Account struct {
	ID          int
	Title       string
	Image       string
	ImageKey    string
	Description string
}
type CartItem struct {
	CartID   int
	CartCost int
	CartName string
}

type Cart struct {
	ID    int
	Items []CartItem
}

func (r *Repository) GetAccount(id int) (Account, error) {
	accounts, err := r.GetAllAccounts()
	if err != nil {
		return Account{}, err
	}

	for _, account := range accounts {
		if account.ID == id {
			return account, nil
		}
	}
	return Account{}, fmt.Errorf("Счёты не найдены")
}
func (r *Repository) GetAllAccounts() ([]Account, error) {
	accounts := []Account{
		{
			ID:          1,
			Title:       "90.01 Выручка",
			Description: "Основной доход от реализации товаров, работ или услуг. Включает выручку от основной деятельности компании без учета налогов и сборов.",
			Image:       "http://127.0.0.1:9000/moneyforecast/revenue.png",
			ImageKey:    "revenue.png",
		},
		{
			ID:          2,
			Title:       "90.02 Себестоимость продаж",
			Description: "Прямые затраты, связанные с производством реализованных товаров или услуг. Включает стоимость сырья, материалов и прямые labor затраты.",
			Image:       "http://127.0.0.1:9000/moneyforecast/cost_of_sales.png",
			ImageKey:    "cost_of_sales.png",
		},
		{
			ID:          3,
			Title:       "91.01 Прочие доходы",
			Description: "Доходы, не связанные с основной деятельностью компании. Включает доходы от аренды, процентов по вкладам, курсовые разницы.",
			Image:       "http://127.0.0.1:9000/moneyforecast/other_income.png",
			ImageKey:    "other_income.png",
		},
		{
			ID:          4,
			Title:       "90.04 Акцизы",
			Description: "Налоги на определенные виды товаров (алкоголь, табак, топливо). Включается в стоимость товара и уплачивается производителем.",
			Image:       "http://127.0.0.1:9000/moneyforecast/excise_taxes.png",
			ImageKey:    "excise_taxes.png",
		},
		{
			ID:          5,
			Title:       "90.09 Прибыль от продаж",
			Description: "Финансовый результат от основной деятельности. Рассчитывается как разница между выручкой и себестоимостью продаж.",
			Image:       "http://127.0.0.1:9000/moneyforecast/sales_profit.png",
			ImageKey:    "sales_profit.png",
		},
		{
			ID:          6,
			Title:       "90.03 Налог на добавленную стоимость",
			Description: "Косвенный налог на добавленную стоимость товаров и услуг. Уплачивается на каждой стадии производства и реализации.",
			Image:       "http://127.0.0.1:9000/moneyforecast/tax.png",
			ImageKey:    "tax.png",
		},
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return accounts, nil
}
func (r *Repository) GetAccountsByTitle(title string) ([]Account, error) {
	accounts, err := r.GetAllAccounts()
	if err != nil {
		return []Account{}, err
	}

	var result []Account
	for _, account := range accounts {
		if strings.Contains(strings.ToLower(account.Title), strings.ToLower(title)) {
			result = append(result, account)
		}
	}
	return result, nil
}
func (r *Repository) GetCart(id int) ([]map[string]interface{}, error) {
	if id != 1 {
		return nil, fmt.Errorf("корзина не найдена")
	}
	cartItems := []CartItem{
		{
			CartID:   1,
			CartName: "90.01 Выручка",
			CartCost: 100,
		},
		{
			CartID:   2,
			CartName: "90.02 Себестоимость продаж",
			CartCost: 500,
		},
		{
			CartID:   6,
			CartName: "90.03 Налог на добавленную стоимость",
			CartCost: 180,
		},
	}
	allAccounts, err := r.GetAllAccounts()
	if err != nil {
		return nil, err
	}
	var Items []map[string]interface{}
	for _, item := range cartItems {
		for _, resource := range allAccounts {
			if resource.ID == item.CartID {
				Item := map[string]interface{}{
					"CartID":              item.CartID,
					"CartAccountImageURL": resource.Image,
					"CartAccountName":     resource.Title,
					"CartAccountCost":     item.CartCost,
				}
				Items = append(Items, Item)
				break
			}
		}
	}
	return Items, nil
}
