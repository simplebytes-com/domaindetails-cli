import importlib.util
import gzip
import sys
import unittest
from pathlib import Path


SCRIPT = Path(__file__).parents[1] / "examples" / "wayback_domain_history.py"
SPEC = importlib.util.spec_from_file_location("wayback_domain_history", SCRIPT)
wayback = importlib.util.module_from_spec(SPEC)
assert SPEC and SPEC.loader
sys.modules[SPEC.name] = wayback
SPEC.loader.exec_module(wayback)


class WaybackDomainHistoryTests(unittest.TestCase):
    def test_normalizes_domain(self):
        self.assertEqual(wayback.normalize_domain("https://WWW.Example.com/path"), "example.com")

    def test_rejects_invalid_domain(self):
        with self.assertRaises(ValueError):
            wayback.normalize_domain("not-a-domain")

    def test_classifies_developed_page(self):
        text = " ".join(["A useful company product and service description"] * 30)
        self.assertEqual(wayback.classify_page(text, "Example Company", 8), ("developed", "high"))

    def test_parked_beats_page_size_signals(self):
        text = "This domain is for sale. " + "word " * 300
        self.assertEqual(wayback.classify_page(text, "Buy this domain", 10), ("parked-or-for-sale", "high"))

    def test_sponsored_directory_is_parked(self):
        text = "Cyclist.com sponsored listings related searches " + "cycling " * 300
        self.assertEqual(wayback.classify_page(text, "Cyclist.com", 50), ("parked-or-for-sale", "high"))

    def test_phpinfo_dump_is_not_developed(self):
        text = "Coming spring of 2017 phpinfo() PHP Version 5.3 System Windows Build Date Jan 2013 " + "config " * 300
        self.assertEqual(
            wayback.classify_page(text, "Cyclist.com - Official Site", 2),
            ("misconfigured-or-placeholder", "high"),
        )

    def test_parser_ignores_script_text_and_collects_links(self):
        parser = wayback.PageParser("https://example.com/")
        parser.feed("<title>Acme</title><script>hidden@example.com</script><p>Email hi@example.com</p><a href='/contact'>Contact</a>")
        text = " ".join(parser.text_parts)
        self.assertNotIn("hidden@example.com", text)
        self.assertIn("hi@example.com", text)
        self.assertEqual(parser.links, ["https://example.com/contact"])

    def test_parser_collects_frame_source(self):
        parser = wayback.PageParser("https://example.com/")
        parser.feed("<frameset><frame src='https://elsewhere.example/site'></frameset>")
        self.assertEqual(parser.frame_sources, ["https://elsewhere.example/site"])
        self.assertEqual(wayback.classify_page("", "", 0, parser.frame_sources), ("frameset-or-forward", "high"))

    def test_parser_decodes_cloudflare_protected_email(self):
        address = "support@example.com"
        key = 0x42
        encoded = f"{key:02x}" + "".join(f"{ord(char) ^ key:02x}" for char in address)
        parser = wayback.PageParser("https://example.com/")
        parser.feed(f'<a data-cfemail="{encoded}">[email protected]</a>')
        self.assertEqual(parser.cloudflare_emails, [address])

    def test_short_numeric_fragment_is_not_a_phone(self):
        self.assertIsNone(wayback.clean_phone("20 30-010"))
        self.assertIsNone(wayback.clean_phone("1023022900919"))
        self.assertIsNone(wayback.clean_phone("172.31.26.189"))
        self.assertEqual(wayback.clean_phone("+357 25 00 00 94"), "+357 25 00 00 94")
        self.assertEqual(wayback.clean_phone("781-839-7903"), "781-839-7903")

    def test_homepage_excludes_query_variants(self):
        self.assertTrue(wayback.is_homepage("https://www.example.com/", "example.com"))
        self.assertFalse(wayback.is_homepage("https://example.com/?add-to-cart=1", "example.com"))

    def test_gzip_fixture_has_expected_magic(self):
        self.assertTrue(gzip.compress(b"<html></html>").startswith(b"\x1f\x8b"))

    def test_timeline_flags_classification_transition(self):
        first = wayback.Snapshot("2020-01-01", "20200101000000", "https://example.com", "replay1", "Acme", "developed", "high", 200, 5, _comparison_text="company products")
        second = wayback.Snapshot("2021-01-01", "20210101000000", "https://example.com", "replay2", "For sale", "parked-or-for-sale", "high", 10, 0, _comparison_text="domain for sale")
        events = wayback.build_timeline([first, second])
        self.assertEqual(events[-1]["type"], "major-change")
        self.assertIn("classification changed", events[-1]["summary"])


if __name__ == "__main__":
    unittest.main()
