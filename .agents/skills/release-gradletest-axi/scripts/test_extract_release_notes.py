import unittest

from extract_release_notes import extract_release_notes


class ExtractReleaseNotesTest(unittest.TestCase):
    def test_extracts_first_release_before_links(self) -> None:
        changelog = """\
# Changelog

## [Unreleased]

## [0.1.0] - 2026-09-20

### Added

- Initial release.

[Unreleased]: https://example.com/compare/v0.1.0...HEAD
[0.1.0]: https://example.com/releases/tag/v0.1.0
"""
        self.assertEqual(
            extract_release_notes(changelog, "v0.1.0"),
            "### Added\n\n- Initial release.\n",
        )

    def test_extracts_later_release_before_previous_release(self) -> None:
        changelog = """\
## [Unreleased]

## [0.2.0] - 2026-10-01

### Changed

- New behavior.

## [0.1.0] - 2026-09-20

### Added

- Initial release.
"""
        self.assertEqual(
            extract_release_notes(changelog, "0.2.0"),
            "### Changed\n\n- New behavior.\n",
        )

    def test_rejects_missing_release(self) -> None:
        with self.assertRaisesRegex(ValueError, "not found"):
            extract_release_notes("## [Unreleased]\n", "v0.1.0")

    def test_rejects_duplicate_release(self) -> None:
        changelog = """\
## [0.1.0] - 2026-09-20

- First.

## [v0.1.0] - 2026-09-21

- Duplicate.
"""
        with self.assertRaisesRegex(ValueError, "duplicate"):
            extract_release_notes(changelog, "v0.1.0")

    def test_rejects_empty_release(self) -> None:
        changelog = """\
## [0.1.0] - 2026-09-20

### Added

[0.1.0]: https://example.com/releases/tag/v0.1.0
"""
        with self.assertRaisesRegex(ValueError, "empty"):
            extract_release_notes(changelog, "v0.1.0")


if __name__ == "__main__":
    unittest.main()
