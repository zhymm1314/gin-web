package fx

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"gin-web/internal/repository"
)

// RepositoryModule 仓储模块
var RepositoryModule = fx.Module("repository",
	fx.Provide(
		ProvideUserRepository,
		ProvideModRepository,
	),
)

// ProvideUserRepository 提供用户仓储
func ProvideUserRepository(db *gorm.DB) repository.UserRepository {
	if db == nil {
		return nil
	}
	return repository.NewUserRepository(db)
}

// ProvideModRepository 提供 Mod 仓储
func ProvideModRepository(db *gorm.DB) repository.ModRepository {
	if db == nil {
		return nil
	}
	return repository.NewModRepository(db)
}
