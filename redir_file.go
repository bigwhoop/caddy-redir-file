package redir_file

import (
	"encoding/csv"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("redir_file", parseCaddyfile)
}

type Redirects map[string]string

// Middleware implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type Middleware struct {
	// Path is the path to a file which contains the redirects
	Path string `json:"path"`

	// Type must be one of: "csv"
	Type string `json:"type"`

	// CsvSeparator allows to overwrite the default CSV separator
	CsvSeparator rune `json:"csv_separator,omitempty"`

	redirects Redirects
	logger    *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.redir_file",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger() // g.logger is a *zap.Logger

	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("failed reading file %s", m.Path)
	}
	defer f.Close()

	switch m.Type {
	case "csv":
		csvReader := csv.NewReader(f)
		csvReader.Comma = m.CsvSeparator

		data, err := csvReader.ReadAll()
		if err != nil {
			return fmt.Errorf("failed reading file %s as CSV", m.Path)
		}

		m.redirects = make(Redirects, len(data))
		for i, line := range data {
			if i == 0 {
				continue // omit header line
			}
			m.redirects[line[0]] = line[1]
		}
		m.logger.Info(fmt.Sprintf("loaded %d redirects from CSV file %s", len(m.redirects), m.Path))
	default:
		return fmt.Errorf("unsupported file type given %s", m.Type)
	}

	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if newPath, ok := m.redirects[r.URL.Path]; ok {
		http.Redirect(w, r, newPath, http.StatusMovedPermanently)
		return nil
	}

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//	redir_file {
//	    path "/var/www/redirects.csv"
//	    type "csv"
//	}
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		if len(args) > 0 {
			return d.ArgErr()
		}

		for d.NextBlock(0) {
			switch d.Val() {
			case "path":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.Path = args[0]
			case "type":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.Type = args[0]
			case "csv_separator":
				args := d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.CsvSeparator = []rune(args[0])[0]
			default:
				return d.Errf("unrecognized subdirective %q", d.Val())
			}
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	m := Middleware{
		CsvSeparator: ',',
	}

	err := m.UnmarshalCaddyfile(h.Dispenser)

	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
