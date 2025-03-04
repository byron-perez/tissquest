package slide

type image struct {
	Url string
}

type staining struct {
	Name string
}

type Slide struct {
	Name          string
	Magnification int
	Staining      staining
	Img           image
}
