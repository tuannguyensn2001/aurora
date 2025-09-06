package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DataType represents the enum for attribute data types
type DataType string

const (
	DataTypeBoolean DataType = "boolean"
	DataTypeString  DataType = "string"
	DataTypeNumber  DataType = "number"
	DataTypeEnum    DataType = "enum"
)

// Attribute represents the attributes table
type Attribute struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"uniqueIndex;not null;size:255" json:"name"`
	Description   string         `gorm:"type:text;not null" json:"description"`
	DataType      DataType       `gorm:"type:data_type;not null;default:'string'" json:"dataType"`
	HashAttribute bool           `gorm:"not null;default:false" json:"hashAttribute"`
	EnumOptions   pq.StringArray `gorm:"type:text[];default:'{}'" json:"enumOptions"`
	UsageCount    int            `gorm:"not null;default:0" json:"usageCount"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (Attribute) TableName() string {
	return "attributes"
}

// BeforeCreate hook to validate enum options
func (a *Attribute) BeforeCreate(tx *gorm.DB) error {
	return a.validate()
}

// BeforeUpdate hook to validate enum options
func (a *Attribute) BeforeUpdate(tx *gorm.DB) error {
	return a.validate()
}

// validate performs validation on the attribute
func (a *Attribute) validate() error {
	// If data type is enum, ensure enum options are provided
	if a.DataType == DataTypeEnum && len(a.EnumOptions) == 0 {
		return gorm.ErrInvalidData
	}
	return nil
}
