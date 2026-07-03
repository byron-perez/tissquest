package migration

import "gorm.io/gorm"

// AnnotationModel stores one W3C Web Annotation per row.
// The full annotation JSON (body, target, creator, etc.) is stored as text
// so we stay format-compatible with the Annotorious standard without
// needing to mirror every nested field as a column.
type AnnotationModel struct {
	gorm.Model
	SlideID        uint           `gorm:"not null;index"`
	AnnotoriousID  string         `gorm:"column:annotorious_id;not null"` // Annotorious-generated UUID
	AnnotationJSON string         `gorm:"column:annotation_json;type:text;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (AnnotationModel) TableName() string {
	return "slide_annotations"
}
