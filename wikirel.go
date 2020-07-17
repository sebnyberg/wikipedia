package wikirel

type FullPageStore interface {
	Get(id int32) (*FullPage, error)
	Set(id int32, p *FullPage) error
}

type FullPageReader interface {
	Next() (*FullPage, error)
}

type FullPageWriter interface {
	Write(*FullPage) error
}
