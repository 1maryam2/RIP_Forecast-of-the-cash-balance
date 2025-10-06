package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"lab_1/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetAccountByID(id int) (*ds.Account, error) {
	query := "SELECT id, code, title, description, is_active FROM accounts WHERE id = $1 and is_active = true"
	row := r.db.Raw(query, id).Row()
	account := &ds.Account{}
	err := row.Scan(
		&account.ID,
		&account.Code,
		&account.Title,
		&account.Description,
		&account.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return account, nil
}

func (r *Repository) CalculateFundsApplicationResult(fundsApplicationID int) (float64, error) {
	var initialSum float64
	query := "SELECT initial_sum FROM funds_applications WHERE id = $1"
	row := r.db.Raw(query, fundsApplicationID).Row()
	if err := row.Scan(&initialSum); err != nil {
		return 0, err
	}
	query = `
        SELECT ci.amount, a.type, a.category 
        FROM funds_application_items ci 
        JOIN accounts a ON ci.account_id = a.id 
        WHERE ci.funds_application_id = $1
    `
	rows, err := r.db.Raw(query, fundsApplicationID).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	result := initialSum

	for rows.Next() {
		var amount float64
		var accType, category string

		if err := rows.Scan(&amount, &accType, &category); err != nil {
			return 0, err
		}
		switch accType {
		case "income":
			result += amount // Доходы увеличивают итог
		case "expense":
			result -= amount // Расходы уменьшают итог
		case "tax":
			result -= amount // Налоги уменьшают итог
		default:
			result -= amount // По умолчанию вычитаем
		}
	}
	return result, nil
}

func (r *Repository) GetAccountIDsInCart(fundsApplicationID uint) ([]uint, error) {
	var accountIDs []uint
	err := r.db.Model(&ds.FundsApplicationItem{}).Where("funds_application_id = ?", fundsApplicationID).Pluck("account_id", &accountIDs).Error
	if err != nil {
		return nil, err
	}
	return accountIDs, nil
}

func (r *Repository) GetOrCreateFundsApplicationForUser(userID uint) (*ds.FundsApplication, error) {
	var fundsApplication ds.FundsApplication
	err := r.db.Where("id = ? AND is_active = ?", 1, true).First(&fundsApplication).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newCart := ds.FundsApplication{
			Name:     fmt.Sprintf("Cart for user %d", userID),
			IsActive: true,
		}
		if err := r.db.Create(&newCart).Error; err != nil {
			return nil, err
		}
		return &newCart, nil
	} else if err != nil {
		return nil, err
	}

	return &fundsApplication, nil
}

func (r *Repository) DeleteFundsApplicationItem(fundsApplicationItemID uint) error {
	err := r.db.Where("id = ?", fundsApplicationItemID).Delete(&ds.FundsApplicationItem{}).Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении элемента корзины: %w", err)
	}
	fmt.Printf("Удален элемент корзины с ID: %d\n", fundsApplicationItemID)
	return nil
}

func (r *Repository) AddToFundsApplication(accountID uint, fundsApplicationID uint) error {
	var existingItems []ds.FundsApplicationItem
	err := r.db.Where("funds_application_id = ? AND account_id = ?", fundsApplicationID, accountID).Find(&existingItems).Error
	if err != nil {
		return err
	}
	if len(existingItems) > 0 {
		return nil
	}
	fundsApplicationItem := ds.FundsApplicationItem{
		FundsApplicationID: fundsApplicationID,
		AccountID:          accountID,
		Amount:             0,
		Comment:            "",
	}
	return r.db.Create(&fundsApplicationItem).Error
}
func (r *Repository) UpdateFundsApplicationResult(fundsApplicationID int, result float64) error {
	err := r.db.Model(&ds.FundsApplication{}).Where("id = ?", fundsApplicationID).Update("result", result).Error
	if err != nil {
		return fmt.Errorf("ошибка при обновлении результата заявки: %w", err)
	}
	return nil
}
func (r *Repository) GetAccount(id int) (ds.Account, error) {
	account := ds.Account{}
	err := r.db.Where("id = ?", id).First(&account).Error
	if err != nil {
		return ds.Account{}, err
	}
	return account, nil
}

func (r *Repository) GetAllAccounts() ([]ds.Account, error) {
	var accounts []ds.Account
	err := r.db.Where("is_active = true").Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return accounts, nil
}

func (r *Repository) GetAccountsByTitle(title string) ([]ds.Account, error) {
	var accounts []ds.Account
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}
func (r *Repository) GetFundsApplicationCount(fundsApplicationID uint) int64 {
	var count int64
	err := r.db.Model(&ds.FundsApplicationItem{}).Where("funds_application_id = ?", fundsApplicationID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in cart_items:", err)
		return 0
	}
	return count
}

func (r *Repository) GetFundsApplication(id int) (map[string]interface{}, error) {
	var fundsApplication ds.FundsApplication
	err := r.db.Preload("FundsApplicationItem.Account").Where("id = ?", id).First(&fundsApplication).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения корзины: %v", err)
	}
	var items []map[string]interface{}
	for _, item := range fundsApplication.FundsApplicationItem {
		items = append(items, map[string]interface{}{
			"FundsApplicationItemID":          item.ID,
			"FundsApplicationID":              item.FundsApplicationID,
			"AccountID":                       item.AccountID,
			"FundsApplicationAccountImageURL": item.Account.Image,
			"FundsApplicationAccountName":     item.Account.Title,
			"FundsApplicationAccountCost":     item.Amount,
			"Comment":                         item.Comment,
		})
	}
	result := map[string]interface{}{
		"FundsApplicationItems":       items,
		"FundsApplicationCompanyName": fundsApplication.CompanyName,
		"FundsApplicationINN":         fundsApplication.INN,
		"FundsApplicationOGRN":        fundsApplication.OGRN,
		"FundsApplicationInitialSum":  fundsApplication.InitialSum,
		"FundsApplicationQuarter":     fundsApplication.Quarter,
		"FundsApplicationResult":      fundsApplication.Result,
	}

	return result, nil
}
