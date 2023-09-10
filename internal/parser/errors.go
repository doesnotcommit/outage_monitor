package parser

type errorParser string

func (e errorParser) Error() string {
	return string(e)
}
func (e errorParser) Parser() {}

const (
	errMapNotFound      errorParser = "map not found"
	errNoRespBody       errorParser = "response body not found"
	errNoOutageStart    errorParser = "outage start not found"
	errNoOutageEnd      errorParser = "outage end not found"
	errNoOutageAffected errorParser = "outage no affected customers"
	errNoAddresses      errorParser = "no addresses"
)
