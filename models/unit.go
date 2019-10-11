package models

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/evilsocket/islazy/log"
	"os"
	"reflect"
	"time"
)

const (
	TokenTTL = time.Minute * 30
)

type Unit struct {
	ID          uint       `gorm:"primary_key" json:"-"`
	CreatedAt   time.Time  `json:"enrolled_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `sql:"index" json:"-"`
	Address     string     `gorm:"size:50;not null" json:"-"`
	Country     string     `gorm:"size:10" json:"country"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Fingerprint string     `gorm:"size:255;not null;unique" json:"fingerprint"`
	PublicKey   string     `gorm:"size:10000;not null" json:"public_key"`
	Token       string     `gorm:"size:10000;not null" json:"-"`
	Data        string     `gorm:"size:10000;not null" json:"data"`

	AccessPoints []AccessPoint `gorm:"foreignkey:UnitID" json:"-"`

	Inbox []Message `gorm:"foreignkey:ReceiverID" json:"-"`
	Sent  []Message `gorm:"foreignkey:SenderID" json:"-"`
}

func (u Unit) Identity() string {
	return fmt.Sprintf("%s@%s", u.Name, u.Fingerprint)
}

func (u Unit) FindAccessPoint(essid, bssid string) *AccessPoint {
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

func (u *Unit) UpdateWith(enroll EnrollmentRequest) error {
	prevData := map[string]interface{}{}

	if u.Data != "" {
		if err := json.Unmarshal([]byte(u.Data), &prevData); err != nil {
			log.Warning("error parsing previous data: %v", err)
			log.Debug("%s", u.Data)
		}
	}

	// only replace sent values
	for key, obj := range enroll.Data {
		set := true
		if key == "session" {
			if session, ok := obj.(map[string]interface{}); !ok {
				set = false
				log.Warning("corrupted session (first level): %v", obj)
			} else if epochs, found := session["epochs"]; !found {
				set = false
				log.Warning("corrupted session (no epochs): %v", obj)
			} else if num, ok := epochs.(float64); !ok {
				set = false
				log.Warning("corrupted session (epochs type %v): %v", reflect.TypeOf(epochs), obj)
			} else if num == 0 {
				// do not update with empty sessions
				set = false
			}
		}

		if set {
			prevData[key] = obj
		}
	}

	newData, err := json.Marshal(prevData)
	if err != nil {
		return err
	}

	if u.Name != enroll.Name {
		log.Info("unit %s changed name: %s -> %s", u.Identity(), u.Name, enroll.Name)
	}

	u.Name = enroll.Name
	u.Address = enroll.Address
	u.Country = enroll.Country
	u.Data = string(newData)

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
	Networks    int                    `json:"networks"`
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
		Networks:    db.Model(u).Association("AccessPoints").Count(),
	}

	if u.Data != "" {
		if err := json.Unmarshal([]byte(u.Data), &doc.Data); err != nil {
			return nil, err
		}
	}

	return json.Marshal(doc)
}
