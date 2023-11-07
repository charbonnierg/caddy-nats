package mongo_input

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (c *MongoInputConnector) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "uri":
			if err := parser.ParseString(d, &c.URI); err != nil {
				return err
			}
		case "collection":
			col := Collection{}
			if err := parser.ParseString(d, &col.Name); err != nil {
				return err
			}
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "database":
					if err := parser.ParseString(d, &col.Database); err != nil {
						return err
					}
				case "pre_and_post":
					if err := parser.ParseBool(d, &col.PreAndPost); err != nil {
						return err
					}
				case "tokens_db_name":
					if err := parser.ParseString(d, &col.TokensDBName); err != nil {
						return err
					}
				case "tokens_collection_name":
					if err := parser.ParseString(d, &col.TokensCollName); err != nil {
						return err
					}
				case "tokens_collection_capped":
					if err := parser.ParseInt64(d, &col.TokensCollCapped); err != nil {
						return err
					}
				case "stream_name":
					if err := parser.ParseString(d, &col.StreamName); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
