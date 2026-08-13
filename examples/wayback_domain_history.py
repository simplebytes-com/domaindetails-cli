#!/usr/bin/env python3
"""Build a small, evidence-linked domain history report from the Wayback Machine.

This is intentionally a local sample rather than a production crawler. It uses
only Python's standard library and is conservative about what it calls a
"developed" website.
"""

from __future__ import annotations

import argparse
import gzip
import html
import json
import re
import sys
import time
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from difflib import SequenceMatcher
from html.parser import HTMLParser
from pathlib import Path
from typing import Iterable
from urllib.error import HTTPError, URLError
from urllib.parse import quote, urljoin, urlparse
from urllib.request import Request, urlopen


CDX_ENDPOINT = "https://web.archive.org/cdx/search/cdx"
REPLAY_ENDPOINT = "https://web.archive.org/web"
USER_AGENT = "DomainDetails-Wayback-History-Sample/0.1 (+https://domaindetails.com)"
EMAIL_RE = re.compile(r"(?<![\w.+-])([\w.+-]+@[\w.-]+\.[A-Za-z]{2,})(?![\w.-])")
PHONE_RE = re.compile(r"(?<!\w)(?:\+?\d[\d\s().-]{6,}\d)(?!\w)")
SPACE_RE = re.compile(r"\s+")
PARKED_RE = re.compile(
    r"\b(domain (?:name )?(?:[\w.-]+ )?is (?:available )?for sale|available for sale|"
    r"buy this domain|we buy domains|domain parking|domainpark|parked (?:free|domain)|"
    r"sponsored listings?|related searches?|make this your home page|anything\.com ltd|"
    r"sedo|afternic|dan\.com|hugedomains|buydomains|bodis)\b",
    re.I,
)
UNDER_CONSTRUCTION_RE = re.compile(
    r"\b(under construction|coming soon|coming (?:spring|summer|fall|winter)|site is being built)\b",
    re.I,
)
MISCONFIGURED_RE = re.compile(r"phpinfo\(\)|PHP Version.{0,200}\bSystem\b.{0,200}\bBuild Date\b", re.I | re.S)


@dataclass(frozen=True)
class Capture:
    timestamp: str
    original: str
    digest: str = ""
    length: int = 0

    @property
    def date(self) -> str:
        return f"{self.timestamp[0:4]}-{self.timestamp[4:6]}-{self.timestamp[6:8]}"

    @property
    def replay_url(self) -> str:
        return f"{REPLAY_ENDPOINT}/{self.timestamp}/{self.original}"

    @property
    def raw_url(self) -> str:
        return f"{REPLAY_ENDPOINT}/{self.timestamp}id_/{self.original}"


@dataclass
class Snapshot:
    date: str
    timestamp: str
    original_url: str
    replay_url: str
    title: str
    classification: str
    confidence: str
    word_count: int
    internal_links: int
    emails: list[str] = field(default_factory=list)
    phones: list[str] = field(default_factory=list)
    social_profiles: list[str] = field(default_factory=list)
    forwards_to: list[str] = field(default_factory=list)
    retrieval_mode: str = "raw"
    text_sample: str = ""
    error: str | None = None
    _comparison_text: str = field(default="", repr=False)

    def public_dict(self) -> dict:
        return {key: value for key, value in asdict(self).items() if not key.startswith("_")}


class PageParser(HTMLParser):
    SKIP = {"script", "style", "svg", "noscript", "template"}

    def __init__(self, base_url: str) -> None:
        super().__init__(convert_charrefs=True)
        self.base_url = base_url
        self.title_parts: list[str] = []
        self.text_parts: list[str] = []
        self.links: list[str] = []
        self.frame_sources: list[str] = []
        self.cloudflare_emails: list[str] = []
        self._skip_depth = 0
        self._in_title = False

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        tag = tag.lower()
        attributes = dict(attrs)
        if tag in self.SKIP:
            self._skip_depth += 1
        if tag == "title":
            self._in_title = True
        if tag == "a":
            href = attributes.get("href")
            if href:
                self.links.append(urljoin(self.base_url, href.strip()))
        if tag in {"frame", "iframe"}:
            source = attributes.get("src")
            if source:
                self.frame_sources.append(urljoin(self.base_url, source.strip()))
        encoded_email = attributes.get("data-cfemail")
        if encoded_email and re.fullmatch(r"[0-9a-fA-F]+", encoded_email):
            try:
                encoded = bytes.fromhex(encoded_email)
                self.cloudflare_emails.append("".join(chr(value ^ encoded[0]) for value in encoded[1:]))
            except (ValueError, IndexError):
                pass

    def handle_endtag(self, tag: str) -> None:
        tag = tag.lower()
        if tag in self.SKIP and self._skip_depth:
            self._skip_depth -= 1
        if tag == "title":
            self._in_title = False

    def handle_data(self, data: str) -> None:
        cleaned = SPACE_RE.sub(" ", html.unescape(data)).strip()
        if not cleaned:
            return
        if self._in_title:
            self.title_parts.append(cleaned)
        if not self._skip_depth:
            self.text_parts.append(cleaned)


def request_text(url: str, timeout: int, retries: int = 2) -> str:
    request = Request(url, headers={"User-Agent": USER_AGENT, "Accept": "text/html,application/json"})
    for attempt in range(retries + 1):
        try:
            with urlopen(request, timeout=timeout) as response:
                charset = response.headers.get_content_charset() or "utf-8"
                body = response.read(5_000_000)
                # Some raw Wayback replays return the originally captured gzip
                # body without a Content-Encoding header.
                if body.startswith(b"\x1f\x8b"):
                    body = gzip.decompress(body)
                return body.decode(charset, errors="replace")
        except HTTPError as exc:
            if exc.code not in {429, 502, 503, 504} or attempt == retries:
                raise
        except (URLError, TimeoutError):
            if attempt == retries:
                raise
        time.sleep(1.5 * (attempt + 1))
    raise RuntimeError("request retries exhausted")


def normalize_domain(value: str) -> str:
    value = value.strip().lower()
    parsed = urlparse(value if "://" in value else f"https://{value}")
    domain = (parsed.hostname or "").rstrip(".")
    if not domain or "." not in domain or not re.fullmatch(r"[a-z0-9.-]+", domain):
        raise ValueError(f"not a valid domain: {value!r}")
    return domain[4:] if domain.startswith("www.") else domain


def cdx_url(domain: str, from_year: int | None, to_year: int | None, limit: int) -> str:
    params = [
        ("url", f"{domain}/*"),
        ("output", "json"),
        ("fl", "timestamp,original,digest,length"),
        ("filter", "statuscode:200"),
        ("filter", "mimetype:text/html"),
        ("collapse", "digest"),
        ("limit", str(limit)),
    ]
    if from_year:
        params.append(("from", str(from_year)))
    if to_year:
        params.append(("to", str(to_year)))
    return CDX_ENDPOINT + "?" + "&".join(f"{quote(k)}={quote(v, safe=':/')}" for k, v in params)


def fetch_captures(domain: str, args: argparse.Namespace) -> list[Capture]:
    captures: list[Capture] = []
    per_host_limit = max(1, args.cdx_limit // 2)
    for host in (domain, f"www.{domain}"):
        payload = json.loads(
            request_text(cdx_url(host, args.from_year, args.to_year, per_host_limit), args.timeout)
        )
        if not payload or len(payload) == 1:
            continue
        headers = payload[0]
        for row in payload[1:]:
            item = dict(zip(headers, row))
            captures.append(
                Capture(
                    timestamp=item["timestamp"],
                    original=item["original"],
                    digest=item.get("digest", ""),
                    length=int(item.get("length") or 0),
                )
            )
    return sorted(set(captures), key=lambda capture: (capture.timestamp, capture.original))


def is_homepage(url: str, domain: str) -> bool:
    parsed = urlparse(url)
    host = (parsed.hostname or "").lower().removeprefix("www.")
    return (
        host == domain
        and not parsed.query
        and parsed.path.rstrip("/") in {"", "/index.html", "/index.htm", "/index.php"}
    )


def choose_captures(captures: list[Capture], domain: str, maximum: int) -> list[Capture]:
    """Prefer homepage change points, then evenly sample if there are too many."""
    candidates = [capture for capture in captures if is_homepage(capture.original, domain)]
    if not candidates:
        candidates = captures
    candidates.sort(key=lambda capture: capture.timestamp)

    unique: list[Capture] = []
    previous_digest = None
    seen_years: set[str] = set()
    for capture in candidates:
        year = capture.timestamp[:4]
        if capture.digest != previous_digest or year not in seen_years:
            unique.append(capture)
            previous_digest = capture.digest
            seen_years.add(year)
    if len(unique) <= maximum:
        return unique
    indexes = {round(i * (len(unique) - 1) / (maximum - 1)) for i in range(maximum)}
    return [capture for index, capture in enumerate(unique) if index in indexes]


def clean_phone(value: str) -> str | None:
    value = SPACE_RE.sub(" ", value).strip(" .,-")
    digits = re.sub(r"\D", "", value)
    if not 9 <= len(digits) <= 15:
        return None
    if value.startswith("+"):
        return value
    if re.fullmatch(r"\d{1,3}(?:\.\d{1,3}){3}", value):
        return None
    if re.search(r"\b\d{4}-\d{2}-\d{2}\b", value):
        return None
    if re.fullmatch(r"\d+[.]\d+\s+\d+[.]\d+", value):
        return None
    separators = len(re.findall(r"[\s().-]", value))
    return value if separators >= 2 else None


def classify_page(text: str, title: str, internal_links: int, frame_sources: list[str] | None = None) -> tuple[str, str]:
    combined = f"{title} {text} {' '.join(frame_sources or [])}"
    words = text.split()
    if PARKED_RE.search(combined):
        return "parked-or-for-sale", "high"
    if MISCONFIGURED_RE.search(combined):
        return "misconfigured-or-placeholder", "high"
    if UNDER_CONSTRUCTION_RE.search(combined):
        return "under-construction", "high"
    if frame_sources and len(words) < 50:
        return "frameset-or-forward", "high"
    strong_signals = sum((len(words) >= 120, internal_links >= 4, len(title) >= 4))
    if strong_signals == 3 or (len(words) >= 250 and internal_links >= 2):
        return "developed", "high"
    if len(words) >= 60 and internal_links >= 2:
        return "developed", "medium"
    if len(words) < 25 and internal_links < 2:
        return "minimal-or-empty", "medium"
    return "uncertain", "low"


def analyze_capture(capture: Capture, domain: str, timeout: int) -> Snapshot:
    try:
        retrieval_mode = "raw"
        try:
            source = request_text(capture.raw_url, timeout)
        except (HTTPError, URLError, TimeoutError):
            # Raw replay occasionally follows an archived redirect back to a
            # now-dead origin. Standard replay is still useful evidence.
            source = request_text(capture.replay_url, timeout)
            retrieval_mode = "standard-replay-fallback"
        parser = PageParser(capture.original)
        parser.feed(source)
        title = SPACE_RE.sub(" ", parser.title_parts[0] if parser.title_parts else "").strip()
        text = SPACE_RE.sub(" ", " ".join(parser.text_parts)).strip()
        internal_links = sum(
            1
            for link in set(parser.links)
            if (urlparse(link).hostname or "").lower().removeprefix("www.") == domain
        )
        emails = set(
            EMAIL_RE.findall(text + " " + " ".join(parser.links) + " " + " ".join(parser.frame_sources))
        ) | set(parser.cloudflare_emails)
        emails = {email for email in emails if not email.lower().endswith((".png", ".jpg", ".gif", ".webp"))}
        phones = {
            phone
            for match in PHONE_RE.findall(text + " " + " ".join(parser.links))
            if (phone := clean_phone(match))
        }
        social_hosts = ("linkedin.com", "twitter.com", "x.com", "facebook.com", "instagram.com", "youtube.com")
        socials = sorted({link for link in parser.links if any(host in (urlparse(link).hostname or "") for host in social_hosts)})
        if retrieval_mode == "standard-replay-fallback" and title.strip().lower() == "wayback machine":
            classification, confidence = "unavailable", "high"
            parser.frame_sources = []
            text = ""
        else:
            classification, confidence = classify_page(text, title, internal_links, parser.frame_sources)
        frame_sources = sorted(
            {
                source
                for source in parser.frame_sources
                if not (urlparse(source).hostname or "").lower().endswith(("archive.org", "web.archive.org"))
            }
        )
        return Snapshot(
            date=capture.date,
            timestamp=capture.timestamp,
            original_url=capture.original,
            replay_url=capture.replay_url,
            title=title,
            classification=classification,
            confidence=confidence,
            word_count=len(text.split()),
            internal_links=internal_links,
            emails=sorted(emails, key=str.lower),
            phones=sorted(phones),
            social_profiles=socials,
            forwards_to=frame_sources,
            retrieval_mode=retrieval_mode,
            text_sample=text[:300],
            _comparison_text=text[:20_000].lower(),
        )
    except (HTTPError, URLError, TimeoutError, UnicodeError) as exc:
        return Snapshot(
            date=capture.date,
            timestamp=capture.timestamp,
            original_url=capture.original,
            replay_url=capture.replay_url,
            title="",
            classification="unavailable",
            confidence="high",
            word_count=0,
            internal_links=0,
            error=str(exc),
        )


def similarity(left: Snapshot, right: Snapshot) -> float:
    if not left._comparison_text or not right._comparison_text:
        return 0.0
    return round(SequenceMatcher(None, left._comparison_text, right._comparison_text).ratio(), 3)


def build_timeline(snapshots: list[Snapshot]) -> list[dict]:
    events: list[dict] = []
    for index, current in enumerate(snapshots):
        if index == 0:
            events.append({"date": current.date, "type": "first-observed", "summary": current.classification})
            continue
        previous = snapshots[index - 1]
        score = similarity(previous, current)
        reasons: list[str] = []
        if previous.classification != current.classification:
            reasons.append(f"classification changed from {previous.classification} to {current.classification}")
        if previous.title != current.title and score < 0.65:
            reasons.append("title and page content changed")
        elif score < 0.35:
            reasons.append("page content changed substantially")
        if reasons:
            events.append(
                {
                    "date": current.date,
                    "previous_observation": previous.date,
                    "type": "major-change",
                    "summary": "; ".join(reasons),
                    "similarity": score,
                    "source": current.replay_url,
                }
            )
    return events


def contact_page_captures(captures: Iterable[Capture], last_developed: Snapshot | None, domain: str) -> list[Capture]:
    if not last_developed:
        return []
    keywords = re.compile(r"/(contact|about|team|impressum|legal|privacy)(?:[/.?#_-]|$)", re.I)
    target_year = int(last_developed.timestamp[:4])
    matching = [
        capture
        for capture in captures
        if keywords.search(urlparse(capture.original).path)
        and abs(int(capture.timestamp[:4]) - target_year) <= 1
        and (urlparse(capture.original).hostname or "").lower().removeprefix("www.") == domain
    ]
    matching.sort(key=lambda item: abs(int(item.timestamp) - int(last_developed.timestamp)))
    result: list[Capture] = []
    seen_urls: set[str] = set()
    for capture in matching:
        normalized = capture.original.lower().rstrip("/")
        if normalized not in seen_urls:
            result.append(capture)
            seen_urls.add(normalized)
        if len(result) == 5:
            break
    return result


def run(args: argparse.Namespace) -> dict:
    domain = normalize_domain(args.domain)
    captures = fetch_captures(domain, args)
    if not captures:
        return {"domain": domain, "generated_at": datetime.now(timezone.utc).isoformat(), "captures_found": 0, "snapshots": []}

    selected = choose_captures(captures, domain, args.max_snapshots)
    snapshots: list[Snapshot] = []
    for index, capture in enumerate(selected, 1):
        print(f"[{index}/{len(selected)}] {capture.date} {capture.original}", file=sys.stderr)
        snapshots.append(analyze_capture(capture, domain, args.timeout))
        if index < len(selected):
            time.sleep(args.delay)

    developed = [snapshot for snapshot in snapshots if snapshot.classification == "developed"]
    last_developed = developed[-1] if developed else None
    contact_snapshots: list[Snapshot] = []
    for capture in contact_page_captures(captures, last_developed, domain):
        contact_snapshots.append(analyze_capture(capture, domain, args.timeout))
        time.sleep(args.delay)

    evidence_classes = {"developed", "parked-or-for-sale", "frameset-or-forward"}
    evidence = [snapshot for snapshot in snapshots if snapshot.classification in evidence_classes]
    evidence.extend(contact_snapshots)
    historical_contacts = {
        "emails": sorted({item for snapshot in evidence for item in snapshot.emails}, key=str.lower),
        "phones": sorted({item for snapshot in evidence for item in snapshot.phones}),
        "social_profiles": sorted({item for snapshot in evidence for item in snapshot.social_profiles}),
        "sources": [
            {"observed_at": snapshot.date, "url": snapshot.replay_url}
            for snapshot in evidence
            if snapshot.emails or snapshot.phones or snapshot.social_profiles
        ],
        "warning": "Historical observations only; ownership and accuracy are not currently verified.",
    }
    return {
        "domain": domain,
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "captures_found": len(captures),
        "snapshots_analyzed": len(snapshots),
        "methodology_note": "Dates are observation dates. A transition happened between adjacent observations, not necessarily on the reported date.",
        "last_developed": last_developed.public_dict() if last_developed else None,
        "historical_contacts": historical_contacts,
        "timeline": build_timeline(snapshots),
        "snapshots": [snapshot.public_dict() for snapshot in snapshots],
        "contact_page_snapshots": [snapshot.public_dict() for snapshot in contact_snapshots],
    }


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("domain", help="Domain or URL to analyze")
    parser.add_argument("--from", dest="from_year", type=int, help="First capture year")
    parser.add_argument("--to", dest="to_year", type=int, help="Last capture year")
    parser.add_argument("--max-snapshots", type=int, default=20, choices=range(2, 101), metavar="2-100")
    parser.add_argument("--cdx-limit", type=int, default=5000, help="Maximum CDX rows to inspect")
    parser.add_argument("--timeout", type=int, default=30, help="HTTP timeout in seconds")
    parser.add_argument("--delay", type=float, default=0.5, help="Delay between replay requests")
    parser.add_argument("--output", type=Path, help="Write JSON to this path instead of stdout")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    try:
        report = run(args)
    except (ValueError, HTTPError, URLError, TimeoutError, json.JSONDecodeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    output = json.dumps(report, indent=2, ensure_ascii=False) + "\n"
    if args.output:
        args.output.write_text(output, encoding="utf-8")
        print(f"Wrote {args.output}", file=sys.stderr)
    else:
        print(output, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
