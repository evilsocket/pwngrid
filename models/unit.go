package models

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/evilsocket/islazy/log"
	"github.com/jinzhu/gorm"
	"os"
	"time"
)

const (
	TokenTTL = time.Minute * 30
)

type Unit struct {
	ID           uint          `gorm:"primary_key" json:"-"`
	CreatedAt    time.Time     `json:"enrolled_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	DeletedAt    *time.Time    `sql:"index" json:"-"`
	Address      string        `gorm:"size:50;not null" json:"-"`
	Country      string        `gorm:"size:10" json:"country"`
	Name         string        `gorm:"size:255;not null" json:"name"`
	Fingerprint  string        `gorm:"size:255;not null;unique" json:"fingerprint"`
	PublicKey    string        `gorm:"size:10000;not null" json:"public_key"`
	Token        string        `gorm:"size:10000;not null" json:"-"`
	Data         string        `gorm:"size:10000;not null" json:"data"`
	AccessPoints []AccessPoint `gorm:"foreignkey:UnitID" json:"-"`
}

func FindUnit(db *gorm.DB, id uint) *Unit {
	var unit Unit
	if err := db.Find(&unit, id).Error; err != nil {
		return nil
	}
	return &unit
}

func FindUnitByFingerprint(db *gorm.DB, fingerprint string) *Unit {
	var unit Unit
	if fingerprint == "" {
		return nil
	} else if err := db.Where("fingerprint = ?", fingerprint).Take(&unit).Error; err != nil {
		return nil
	}
	return &unit
}

func EnrollUnit(db *gorm.DB, enroll EnrollmentRequest) (err error, unit *Unit) {
	if unit = FindUnitByFingerprint(db, enroll.Fingerprint); unit == nil {
		log.Info("enrolling new unit for %s (%s): %s", enroll.Address, enroll.Country, enroll.Identity)

		unit = &Unit{
			Address:     enroll.Address,
			Country:     enroll.Country,
			Name:        enroll.Name,
			Fingerprint: enroll.Fingerprint,
			PublicKey:   string(enroll.KeyPair.PublicPEM),
		}

		if err := db.Create(unit).Error; err != nil {
			return fmt.Errorf("error enrolling %s: %v", unit.Identity(), err), nil
		}
	}

	if err := unit.updateToken(); err != nil {
		return fmt.Errorf("error creating token for %s: %v", unit.Identity(), err), nil
	}

	if err = unit.UpdateWith(db, enroll); err != nil {
		return fmt.Errorf("error setting token for %s: %v", unit.Identity(), err), nil
	}
	return nil, unit
}

func (u Unit) Identity() string {
	return fmt.Sprintf("%s@%s", u.Name, u.Fingerprint)
}

func (u Unit) FindAccessPoint(db *gorm.DB, essid, bssid string) *AccessPoint {
	var ap AccessPoint

	if err := db.Where("unit_id = ? AND name = ? AND mac = ?", u.ID, essid, bssid).Take(&ap).Error; err != nil {
		if err := db.Where("unit_id = ? AND mac = ?", u.ID, bssid).Take(&ap).Error; err != nil {
			return nil
		}
	}

	return &ap
}

func (u *Unit) updateToken() error {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["unit_id"] = u.ID
	claims["unit_ident"] = u.Identity()
	claims["expires_at"] = time.Now().Add(TokenTTL).Format(time.RFC3339)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("API_SECRET")))
	if err != nil {
		return err
	}
	u.Token = signed
	return nil
}

func (u *Unit) UpdateWith(db *gorm.DB, enroll EnrollmentRequest) error {
	data, err := json.Marshal(enroll.Data)
	if err != nil {
		return err
	}

	if u.Name != enroll.Name {
		log.Info("unit %s changed name: %s -> %s", u.Identity(), u.Name, enroll.Name)
	}

	u.Name = enroll.Name
	u.Address = enroll.Address
	u.Country = enroll.Country
	u.Data = string(data)

	return db.Save(u).Error
}

type unitJSON struct {
	EnrolledAt  time.Time              `json:"enrolled_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Country     string                 `json:"country"`
	Name        string                 `json:"name"`
	Fingerprint string                 `json:"fingerprint"`
	PublicKey   string                 `json:"public_key"`
	Data        map[string]interface{} `json:"data"`
}

func (u *Unit) MarshalJSON() ([]byte, error) {
	doc := unitJSON{
		EnrolledAt:  u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Country:     u.Country,
		Name:        u.Name,
		Fingerprint: u.Fingerprint,
		PublicKey:   u.PublicKey,
		Data:        map[string]interface{}{},
	}

	if u.Data != "" {
		if err := json.Unmarshal([]byte(u.Data), &doc.Data); err != nil {
			return nil, err
		}
	}

	return json.Marshal(doc)
}
