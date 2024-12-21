package utils

import "gorm.io/gorm"

func TransactionPanicRecover(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	} else if tx.Error != nil {
		tx.Rollback()
	}
}
