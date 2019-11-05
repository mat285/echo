package types

// DomainNameEntry is an entry for a domain name owner
type DomainNameEntry struct {
	Service string `json:"service"`
	Domain  string `json:"domain"`
}
