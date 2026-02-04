package repository

import (
	"gin-web/app/models"
	"gorm.io/gorm"
)

// ModRepository Mod仓储接口
type ModRepository interface {
	Search(db *gorm.DB) ([]models.Mod, error)
	Count(db *gorm.DB) (int64, error)
	FindByID(id uint) (*models.Mod, error)
	UpdateDownloadCount(mod *models.Mod) error
	FindAllGames() ([]models.Game, error)
	FindAllCategories() ([]models.Category, error)
}

type modRepository struct {
	db *gorm.DB
}

// NewModRepository 创建 Mod 仓储实例
func NewModRepository(db *gorm.DB) ModRepository {
	return &modRepository{db: db}
}

func (r *modRepository) Search(query *gorm.DB) ([]models.Mod, error) {
	var mods []models.Mod
	if err := query.Find(&mods).Error; err != nil {
		return nil, err
	}
	return mods, nil
}

func (r *modRepository) Count(query *gorm.DB) (int64, error) {
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *modRepository) FindByID(id uint) (*models.Mod, error) {
	var mod models.Mod
	if err := r.db.Preload("Game").Preload("Categories").First(&mod, id).Error; err != nil {
		return nil, err
	}
	return &mod, nil
}

func (r *modRepository) UpdateDownloadCount(mod *models.Mod) error {
	return r.db.Model(mod).UpdateColumn("download_count", mod.DownloadCount+1).Error
}

func (r *modRepository) FindAllGames() ([]models.Game, error) {
	var games []models.Game
	if err := r.db.Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (r *modRepository) FindAllCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
