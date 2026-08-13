# Changelog

## Unreleased

- Add a dependency-free local Wayback Machine domain-history sample that detects major historical changes, estimates the last developed capture, and extracts evidence-linked historical contact details.
- Decode raw archived gzip responses, exclude query-string variants from homepage sampling, and recognize historical frameset forwarding pages.
- Preserve historical frame destinations, reject short phone-number false positives, and retry failed raw captures through standard Wayback replay.
- Decode Cloudflare-protected email addresses embedded in archived contact pages.
- Distinguish sponsored domain directories and exposed server diagnostics from developed websites, and retain contact evidence from historical parked and forwarding phases.
- Reject bare numeric IDs, dates, IP addresses, and coordinate-like values from historical phone extraction.
