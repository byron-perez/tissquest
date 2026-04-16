package taxon

import "errors"

var (
	ErrEmptyName   = errors.New("taxon name must not be empty")
	ErrInvalidRank = errors.New("taxon rank is not valid")
)

type Rank string

const (
	RankKingdom Rank = "kingdom"
	RankPhylum  Rank = "phylum"
	RankClass   Rank = "class"
	RankOrder   Rank = "order"
	RankFamily  Rank = "family"
	RankGenus   Rank = "genus"
	RankSpecies Rank = "species"
)

type Taxon struct {
	ID       uint
	Rank     Rank
	Name     string
	ParentID *uint
	Parent   *Taxon
}

// Validate checks that the taxon has a non-empty name and a valid rank.
func (t *Taxon) Validate() error {
	if t.Name == "" {
		return ErrEmptyName
	}
	switch t.Rank {
	case RankKingdom, RankPhylum, RankClass, RankOrder, RankFamily, RankGenus, RankSpecies:
		return nil
	default:
		return ErrInvalidRank
	}
}

// Lineage returns the full classification path from root to this taxon.
func (t *Taxon) Lineage() []Taxon {
	if t.Parent == nil {
		return []Taxon{*t}
	}
	return append(t.Parent.Lineage(), *t)
}

type RepositoryInterface interface {
	Save(t *Taxon) (uint, error)
	GetByID(id uint) (*Taxon, error)
	Update(id uint, t *Taxon) error
	GetLineage(id uint) ([]Taxon, error)
	ListByRank(rank Rank) ([]Taxon, error)
	List() ([]Taxon, error)
	Delete(id uint) error
}
