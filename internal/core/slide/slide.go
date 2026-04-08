package slide

type Preparation struct {
	Staining        string
	InclusionMethod string
	Reagents        string
	Protocol        string
	Notes           string
}

type Slide struct {
	Name          string
	Url           string
	Magnification int
	Preparation   Preparation
}
