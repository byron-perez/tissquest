package migration

import "gorm.io/gorm"

// CollectionModel maps to the "collections" table.
type CollectionModel struct {
	gorm.Model
	Name        string
	Description string
	Goals       string
	Type        string // "atlas" | "database" | "reference" | "other"
	Authors     string
	Sections    []CollectionSectionModel `gorm:"foreignKey:CollectionID"`
}

func (CollectionModel) TableName() string {
	return "collections"
}

// CollectionSectionModel maps to the "collection_sections" table.
type CollectionSectionModel struct {
	gorm.Model
	CollectionID uint
	ParentID     *uint
	Name         string
	Position     int
	Assignments  []CollectionSectionAssignmentModel `gorm:"foreignKey:SectionID"`
	Subsections  []CollectionSectionModel           `gorm:"foreignKey:ParentID"`
}

func (CollectionSectionModel) TableName() string {
	return "collection_sections"
}

// CollectionSectionAssignmentModel maps to the "collection_section_assignments" table.
type CollectionSectionAssignmentModel struct {
	gorm.Model
	SectionID      uint
	TissueRecordID uint
	TissueRecord   TissueRecordModel `gorm:"foreignKey:TissueRecordID"`
	Position       int
}

func (CollectionSectionAssignmentModel) TableName() string {
	return "collection_section_assignments"
}
