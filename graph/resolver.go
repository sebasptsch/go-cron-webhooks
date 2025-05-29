package graph

import (
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB         *gorm.DB
	C          *cron.Cron
	WebhookMap map[uint]cron.EntryID
}
