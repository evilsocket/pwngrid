package models

import "github.com/biezhi/gorm-paginator/pagination"

type UnitsByCountry struct {
	Country string `json:"country"`
	Count   int    `json:"units"`
}

func GetUnitsByCountry() ([]UnitsByCountry, error) {
	results := make([]UnitsByCountry, 0)
	if err := db.Raw("SELECT country,COUNT(id) AS count FROM units GROUP BY country ORDER BY count DESC").Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func GetPagedUnits(page int) (units []Unit, total int, pages int) {
	paginator := pagination.Paging(&pagination.Param{
		DB:      db,
		Page:    page,
		Limit:   25,
		OrderBy: []string{"id desc"},
	}, &units)
	return units, paginator.TotalRecord, paginator.TotalPage
}

func FindUnit(id uint) *Unit {
	var unit Unit
	if err := db.Find(&unit, id).Error; err != nil {
		return nil
	}
	return &unit
}

func FindUnitByFingerprint(fingerprint string) *Unit {
	var unit Unit
	if fingerprint == "" {
		return nil
	} else if err := db.Where("fingerprint = ?", fingerprint).Take(&unit).Error; err != nil {
		return nil
	}
	return &unit
}
