package persistence

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"remi-jadul/server/internal/auth"
)

var (
	ErrDatabase      = errors.New("persistence: database error")
	ErrGuestNotFound = auth.ErrGuestNotFound
)

type GuestModel struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	DisplayName string    `gorm:"column:display_name;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

func (GuestModel) TableName() string { return "guests" }

type GuestRepository struct {
	db *gorm.DB
}

// Alias types.
type GuestStore = GuestRepository
type PostgresGuestRepository = GuestRepository
type GuestPersistence = GuestRepository

func NewGuestRepository(db *gorm.DB) *GuestRepository {
	return &GuestRepository{db: db}
}

func NewPostgresGuestRepository(databaseURL string) (*GuestRepository, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return NewGuestRepository(db), nil
}

// We implement opener via driver import side-effect only; import driver in file.
// Adjust imports in separate file.
func NewGuestRepositoryFromDB(db *gorm.DB) *GuestRepository { return NewGuestRepository(db) }
func NewGuestRepo(db *gorm.DB) *GuestRepository             { return NewGuestRepository(db) }

func (r *GuestRepository) DB() *gorm.DB { return r.db }

func (r *GuestRepository) CreateGuest(displayName string) (*auth.Guest, error) {
	trimmed := strings.TrimSpace(displayName)
	if trimmed == "" {
		return nil, auth.ErrBlankNickname
	}
	now := time.Now().UTC()
	model := GuestModel{
		ID:          newUUID(),
		DisplayName: trimmed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.db.Create(&model).Error; err != nil {
		return nil, fmtError(err)
	}
	return &auth.Guest{
		ID:          model.ID,
		DisplayName: model.DisplayName,
		Nickname:    model.DisplayName,
		Name:        model.DisplayName,
	}, nil
}

func (r *GuestRepository) GetGuest(id string) (*auth.Guest, error) {
	var model GuestModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrGuestNotFound
		}
		return nil, fmtError(err)
	}
	return &auth.Guest{
		ID:          model.ID,
		DisplayName: model.DisplayName,
		Nickname:    model.DisplayName,
		Name:        model.DisplayName,
	}, nil
}

func (r *GuestRepository) DeleteGuest(id string) error {
	res := r.db.Delete(&GuestModel{}, "id = ?", id)
	if res.Error != nil {
		return fmtError(res.Error)
	}
	if res.RowsAffected == 0 {
		return auth.ErrGuestNotFound
	}
	return nil
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexStr := hex.EncodeToString(b)
	return hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" + hexStr[16:20] + "-" + hexStr[20:32]
}

func fmtError(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrDatabase, err)
}
