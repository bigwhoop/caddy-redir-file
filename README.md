# caddy-redir-file

A caddy plugin to load a list of redirects from a file.
It's very simple. And I've only hacked it together in a few hours.
I'd not advise you to use it in production ...

## Installation

Build caddy with `xcaddy`:

```bash
xcaddy build --with github.com/bigwhoop/caddy-redir-file
```

## Configuration

#### Caddyfile
```
{
    order redir_file first
}

www.domain.com {
    redir_file {
        path "/path/to/redirects.csv"
        type "csv"
    }
}
```

#### /path/to/redirects.csv
```
from;to
/from-here;/to-here
/and-here;/to-here
```

## Development

```bash
cd <repo>
xcaddy build --with github.com/bigwhoop/caddy-redir-file=./ --output caddy
./caddy run --config example/Caddyfile
```

Visit [http://localhost:8090/from/here](http://localhost:8090/from/here). 

## License

MIT