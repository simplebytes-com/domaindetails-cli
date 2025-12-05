# domaindetails

A fast CLI tool for domain registration lookups using RDAP and WHOIS.

[![Docker Hub](https://img.shields.io/docker/v/domaindetails/cli?label=docker)](https://hub.docker.com/r/domaindetails/cli)
[![Homebrew](https://img.shields.io/badge/homebrew-domaindetails-blue)](https://github.com/simplebytes-com/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Built by [DomainDetails.com](https://domaindetails.com) - Privacy-first domain intelligence.

## Features

- **RDAP Lookups** - Modern JSON-based protocol (preferred)
- **WHOIS Fallback** - Traditional protocol via DomainDetails.com API
- **Local Caching** - Caches IANA bootstrap data for fast TLD resolution
- **Multiple Output Formats** - Human-readable, JSON, or raw data
- **Cross-Platform** - macOS, Linux, Windows

## Installation

### Homebrew (macOS/Linux)

```bash
brew install simplebytes-com/tap/domaindetails
```

### Docker

```bash
docker pull domaindetails/cli

# Run a lookup
docker run domaindetails/cli lookup example.com
```

### Binary Download

Download the latest release for your platform from [GitHub Releases](https://github.com/simplebytes-com/domaindetails-cli/releases).

### Build from Source

```bash
git clone https://github.com/simplebytes-com/domaindetails-cli.git
cd domaindetails-cli
make build
make install
```

## Usage

### Basic Lookup

```bash
# Auto-select best method (RDAP preferred, WHOIS fallback)
domaindetails lookup example.com

# RDAP only
domaindetails rdap example.com

# WHOIS only
domaindetails whois example.com
```

### Output Formats

```bash
# JSON output
domaindetails lookup example.com --json

# Include raw response data
domaindetails lookup example.com --raw

# JSON with raw data
domaindetails lookup example.com --json --raw
```

### Verbose Mode

```bash
domaindetails lookup example.com --verbose
```

### Cache Management

```bash
# Update the IANA bootstrap cache
domaindetails cache update

# View cache status
domaindetails cache info

# Clear the cache
domaindetails cache clear
```

## Example Output

```
────────────────────────────────────────────────────────────
Domain: example.com
Method: RDAP
────────────────────────────────────────────────────────────

Domain Name:     EXAMPLE.COM
Registrar:       RESERVED-Internet Assigned Numbers Authority

Created:         1995-08-14T04:00:00Z
Expires:         2025-08-13T04:00:00Z
Last Modified:   2024-08-14T07:01:34Z

Status:
  • client delete prohibited
  • client transfer prohibited
  • client update prohibited

Nameservers:
  • A.IANA-SERVERS.NET
  • B.IANA-SERVERS.NET

DNSSEC:          signed
```

### JSON Output

```json
{
  "domain": "example.com",
  "available": false,
  "method": "rdap",
  "parsed": {
    "domainName": "EXAMPLE.COM",
    "registrar": "RESERVED-Internet Assigned Numbers Authority",
    "creationDate": "1995-08-14T04:00:00Z",
    "expirationDate": "2025-08-13T04:00:00Z",
    "lastModified": "2024-08-14T07:01:34Z",
    "nameservers": ["A.IANA-SERVERS.NET", "B.IANA-SERVERS.NET"],
    "status": ["client delete prohibited", "client transfer prohibited"],
    "dnssec": "signed"
  }
}
```

## How It Works

1. **RDAP Lookups**: Queries RDAP servers directly using the IANA bootstrap file (cached locally at `~/.domaindetails/`)
2. **WHOIS Lookups**: Routes through the [DomainDetails.com API](https://api.domaindetails.io) which handles raw WHOIS queries and parsing

### IANA Bootstrap Cache

The CLI caches the IANA RDAP bootstrap file locally to avoid repeated network requests:

- **Location**: `~/.domaindetails/rdap-bootstrap.json`
- **TTL**: 24 hours
- **Fallback**: Uses stale cache if refresh fails

## API

This CLI uses the DomainDetails.com API for WHOIS lookups:

| Endpoint | Description |
|----------|-------------|
| `GET /api/whois?domain=example.com` | WHOIS lookup with parsing |

The WHOIS parser is open source: [@domaindetails/whois-parser](https://www.npmjs.com/package/@domaindetails/whois-parser)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- **Website**: [domaindetails.com](https://domaindetails.com)
- **Documentation**: [domaindetails.com/kb/cli](https://domaindetails.com/kb/cli)
- **Docker Hub**: [hub.docker.com/r/domaindetails/cli](https://hub.docker.com/r/domaindetails/cli)
- **WHOIS Parser**: [npmjs.com/package/@domaindetails/whois-parser](https://www.npmjs.com/package/@domaindetails/whois-parser)
