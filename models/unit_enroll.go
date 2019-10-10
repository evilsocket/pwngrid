package models

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
)

func EnrollUnit(enroll EnrollmentRequest) (err error, unit *Unit) {
	if unit = FindUnitByFingerprint(enroll.Fingerprint); unit == nil {
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

	if err = unit.UpdateWith(enroll); err != nil {
		log.Debug("%+v", enroll)
		return fmt.Errorf("error setting token for %s: %v", unit.Identity(), err), nil
	}
	return nil, unit
}
