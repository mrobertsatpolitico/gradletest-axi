#!/usr/bin/env python3

import argparse
import re
import sys
from pathlib import Path


SEMVER_PATTERN = re.compile(
    r"^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)"
    r"(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?"
    r"(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$"
)
VERSION_HEADING_PATTERN = re.compile(
    r"^## \[(?P<version>[^\]]+)\](?: - \d{4}-\d{2}-\d{2})?\s*$"
)
REFERENCE_PATTERN = re.compile(r"^\[[^\]]+\]:\s+\S+")


def normalize_version(version: str) -> str:
    if not SEMVER_PATTERN.fullmatch(version):
        raise ValueError(f"invalid semantic version: {version}")
    return version.removeprefix("v")


def extract_release_notes(changelog: str, requested_version: str) -> str:
    target = normalize_version(requested_version)
    lines = changelog.splitlines()
    matches: list[int] = []

    for index, line in enumerate(lines):
        match = VERSION_HEADING_PATTERN.fullmatch(line)
        if match is None or match.group("version") == "Unreleased":
            continue
        if normalize_version(match.group("version")) == target:
            matches.append(index)

    if not matches:
        raise ValueError(f"release section not found: {requested_version}")
    if len(matches) > 1:
        raise ValueError(f"duplicate release sections: {requested_version}")

    start = matches[0] + 1
    end = len(lines)
    for index in range(start, len(lines)):
        if VERSION_HEADING_PATTERN.fullmatch(lines[index]):
            end = index
            break
        if REFERENCE_PATTERN.fullmatch(lines[index]):
            end = index
            break

    section_lines = lines[start:end]
    while section_lines and not section_lines[0].strip():
        section_lines.pop(0)
    while section_lines and not section_lines[-1].strip():
        section_lines.pop()

    content_lines = [
        line
        for line in section_lines
        if line.strip() and not line.startswith("### ")
    ]
    if not content_lines:
        raise ValueError(f"release section is empty: {requested_version}")

    return "\n".join(section_lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Extract one version body from a Keep a Changelog file."
    )
    parser.add_argument("changelog", type=Path)
    parser.add_argument("version")
    args = parser.parse_args()

    try:
        changelog = args.changelog.read_text(encoding="utf-8")
        notes = extract_release_notes(changelog, args.version)
    except (OSError, ValueError) as error:
        print(f"extract release notes: {error}", file=sys.stderr)
        return 2

    sys.stdout.write(notes)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
