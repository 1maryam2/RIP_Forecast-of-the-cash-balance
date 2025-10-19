package repository

import (
	"errors"
	"fmt"
	"io"
	"lab_1/internal/app/ds"
	"math/rand"
	"time"

	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MockMinioClient struct{}

func (m *MockMinioClient) PutObject(bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	fmt.Printf("MinIO Mock: Uploading %s/%s\n", bucketName, objectName)
	return minio.UploadInfo{Key: objectName}, nil
}

func (m *MockMinioClient) RemoveObject(bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	fmt.Printf("MinIO Mock: Deleting %s/%s\n", bucketName, objectName)
	return nil
}

func (r *Repository) CreateUser(user *ds.Users) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByLogin(login string) (*ds.Users, error) {
	var user ds.Users
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uint) (*ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (r *Repository) UpdateUser(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ds.Users{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateAccount(account *ds.Account) error {
	return r.db.Create(account).Error
}

func (r *Repository) UpdateAccount(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ds.Account{}).Where("id = ? AND is_active = true", id).Updates(updates).Error
}

func (r *Repository) DeleteAccount(id uint) error {
	var account ds.Account
	if err := r.db.Where("id = ? AND is_active = true", id).First(&account).Error; err != nil {
		return err
	}
	err := r.db.Model(&ds.Account{}).Where("id = ?", id).Update("is_active", false).Error
	if err != nil {
		return err
	}
	if account.Image != "" {
		//r.MinIOClient.RemoveObject(AccountBucket, account.Image, minio.RemoveObjectOptions{})
		fmt.Printf("MinIO Mock: Image removed for account %d: %s\n", id, account.Image)
	}

	return nil
}

func (r *Repository) SetAccountImage(accountID uint, imageName string) error {
	return r.db.Model(&ds.Account{}).Where("id = ? AND is_active = true", accountID).Update("image", imageName).Error
}

func (r *Repository) GetAllAccounts() ([]ds.Account, error) {
	var accounts []ds.Account
	err := r.db.Where("is_active = true").Find(&accounts).Error
	return accounts, err
}

func (r *Repository) GetAccountsByTitleByFilter(query ds.AccountsFilter) ([]ds.Account, error) {
	var accounts []ds.Account
	dbQuery := r.db.Where("is_active = true")

	if query.Search != "" {
		dbQuery = dbQuery.Where("title ILIKE ?", "%"+query.Search+"%")
	}
	if query.Type != "" {
		dbQuery = dbQuery.Where("type = ?", query.Type)
	}
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	err := dbQuery.Find(&accounts).Error
	return accounts, err
}

func (r *Repository) GetAccountByID(id uint) (*ds.Account, error) {
	var account ds.Account
	err := r.db.Where("id = ? AND is_active = true", id).First(&account).Error
	return &account, err
}

func (r *Repository) GetCashForecastCountForUser(userID uint) (uint, int64, error) {
	draft, err := r.GetUserDraftCashForecast(userID)
	if err != nil {
		return 0, 0, err
	}

	var count int64
	err = r.db.Model(&ds.FundsApplicationItem{}).
		Where("funds_application_id = ?", draft.ID).
		Count(&count).Error

	return draft.ID, count, err
}

func (r *Repository) GetCashForecastsList(filter ds.ApplicationFilter) ([]ds.FundsApplication, error) {
	var applications []ds.FundsApplication
	dbQuery := r.db.Where("is_active = true")

	if filter.DateFrom != nil && filter.DateTo != nil {
		dbQuery = dbQuery.Where("formed_at BETWEEN ? AND ?", *filter.DateFrom, *filter.DateTo)
	}

	err := dbQuery.
		Preload("Creator").
		Preload("Moderator").
		Find(&applications).Error
	for i := range applications {
		if applications[i].Creator != nil {
			applications[i].Creator.Password = ""
		}
		if applications[i].Moderator != nil {
			applications[i].Moderator.Password = ""
		}
	}

	return applications, err
}

func (r *Repository) GetUserDraftCashForecast(userID uint) (*ds.FundsApplication, error) {
	var application ds.FundsApplication
	err := r.db.Where("creator_id = ? AND status = ? AND is_active = true", userID, ds.Draft).First(&application).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newApp := ds.FundsApplication{
			Name:        fmt.Sprintf("Прогноз остатка для пользователя %d", userID),
			CreatorID:   userID,
			Status:      ds.Draft,
			IsActive:    true,
			InitialSum:  1000.0,
			CompanyName: "Название компании",
			INN:         "1234567890",
			OGRN:        "1234567890123",
			Quarter:     1,
		}
		if err := r.db.Create(&newApp).Error; err != nil {
			return nil, err
		}
		return &newApp, nil
	}
	return &application, err
}

func (r *Repository) GetCashForecastByID(id uint) (*ds.FundsApplication, error) {
	var application ds.FundsApplication
	err := r.db.Preload("FundsApplicationItem.Account").
		Preload("Creator").
		Preload("Moderator").
		Where("is_active = true").
		First(&application, id).Error

	if application.Creator != nil {
		application.Creator.Password = ""
	}
	if application.Moderator != nil {
		application.Moderator.Password = ""
	}

	return &application, err
}

func (r *Repository) AddAccountToCashForecast(applicationID, accountID uint) error {
	var count int64
	r.db.Model(&ds.FundsApplicationItem{}).Where("funds_application_id = ? AND account_id = ?", applicationID, accountID).Count(&count)
	if count > 0 {
		return errors.New("account already in application")
	}
	var app ds.FundsApplication
	if err := r.db.Select("status").First(&app, applicationID).Error; err != nil {
		return err
	}
	if app.Status != ds.Draft {
		return errors.New("cannot add items to non-draft application")
	}

	item := ds.FundsApplicationItem{
		FundsApplicationID: applicationID,
		AccountID:          accountID,
		Amount:             1.0,
	}
	return r.db.Create(&item).Error
}

func (r *Repository) RemoveItemFromCashForecast(itemID uint) error {
	var item ds.FundsApplicationItem
	if err := r.db.Preload("FundsApplication").First(&item, itemID).Error; err != nil {
		return err
	}
	if item.FundsApplication.Status != ds.Draft {
		return errors.New("cannot remove item from non-draft application")
	}

	return r.db.Delete(&ds.FundsApplicationItem{}, itemID).Error
}

func (r *Repository) UpdateCashForecastItem(itemID uint, req ds.UpdateFundsApplicationItemRequest) error {
	var item ds.FundsApplicationItem
	if err := r.db.Preload("FundsApplication").First(&item, itemID).Error; err != nil {
		return err
	}
	if item.FundsApplication.Status != ds.Draft {
		return errors.New("cannot update item in non-draft application")
	}

	return r.db.Model(&ds.FundsApplicationItem{}).Where("id = ?", itemID).Updates(req).Error
}

func (r *Repository) UpdateCashForecast(appID uint, req ds.UpdateFundsApplicationRequest) error {
	var app ds.FundsApplication
	if err := r.db.Select("status").First(&app, appID).Error; err != nil {
		return err
	}
	if app.Status != ds.Draft {
		return errors.New("only draft applications can be updated")
	}

	updates := map[string]interface{}{}
	if req.CompanyName != "" {
		updates["company_name"] = req.CompanyName
	}
	if req.INN != "" {
		updates["inn"] = req.INN
	}
	if req.OGRN != "" {
		updates["ogrn"] = req.OGRN
	}
	if req.InitialSum != 0 {
		updates["initial_sum"] = req.InitialSum
	}
	if req.Quarter != 0 {
		updates["quarter"] = req.Quarter
	}

	return r.db.Model(&ds.FundsApplication{}).Where("id = ?", appID).Updates(updates).Error
}
func (r *Repository) FormCashForecast(appID uint) error {
	var app ds.FundsApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return err
	}
	if app.Status != ds.Draft {
		return errors.New("only draft application can be formed")
	}
	if app.CompanyName == "" || app.INN == "" || app.OGRN == "" || app.InitialSum <= 0 || app.Quarter == 0 {
		return errors.New("application is missing required fields (company_name, inn, ogrn, initial_sum, quarter)")
	}
	var count int64
	r.db.Model(&ds.FundsApplicationItem{}).Where("funds_application_id = ?", appID).Count(&count)
	if count == 0 {
		return errors.New("application must contain at least one account item")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":    ds.Formed,
		"formed_at": &now,
	}
	return r.db.Model(&ds.FundsApplication{}).Where("id = ?", appID).Updates(updates).Error
}
func calculateResult(initialSum float64, quarter int) float64 {
	bonus := rand.Float64() * 1000
	return initialSum*(1+float64(quarter)*0.05) + bonus
}

func (r *Repository) CompleteCashForecast(appID uint, moderatorID uint) error {
	var app ds.FundsApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return err
	}
	if app.Status != ds.Formed {
		return errors.New("only formed application can be completed")
	}
	calculatedResult := calculateResult(app.InitialSum, app.Quarter)

	now := time.Now()
	updates := map[string]interface{}{
		"status":       ds.Completed,
		"moderator_id": moderatorID,
		"completed_at": &now,
		"result":       calculatedResult,
	}
	return r.db.Model(&ds.FundsApplication{}).Where("id = ?", appID).Updates(updates).Error
}
func (r *Repository) RejectCashForecast(appID uint, moderatorID uint) error {
	var app ds.FundsApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return err
	}
	if app.Status != ds.Formed {
		return errors.New("only formed application can be rejected")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       ds.Rejected,
		"moderator_id": moderatorID,
		"completed_at": &now,
	}
	return r.db.Model(&ds.FundsApplication{}).Where("id = ?", appID).Updates(updates).Error
}

func (r *Repository) DeleteCashForecast(appID uint) error {
	var app ds.FundsApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return err
	}
	if app.Status != ds.Draft {
		return errors.New("only draft applications can be deleted")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":    ds.Deleted,
		"is_active": false,
		"formed_at": &now,
	}
	return r.db.Model(&ds.FundsApplication{}).Where("id = ?", appID).Updates(updates).Error
}

var MockMinIOClient = &MockMinioClient{}

func (r *Repository) SetMinIOClient(client interface{}) {
	fmt.Println("MinIO Client set (mocked)")
}
