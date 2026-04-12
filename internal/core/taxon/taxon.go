package taxon

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
	GetLineage(id uint) ([]Taxon, error)
	ListByRank(rank Rank) ([]Taxon, error)
}
