package session

/*
Cookie
source the source url
name the cookie name
value the cookie value
domain the cookie domain
path the cookie path
creation the creation date
expiry the expiration date
last_access the last time the cookie was accessed
secure is the cookie secure
http_only is the cookie only valid for HTTP
priority internal priority information
*/
type Cookie struct {
	Source     string `json:"source"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	Creation   int64  `json:"creation"`
	Expiry     int64  `json:"expiry"`
	LastAccess int64  `json:"last_access"`
	Secure     uint   `json:"secure"`
	HttpOnly   bool   `json:"http_only"`
	Priority   uint   `json:"priority"`
}
