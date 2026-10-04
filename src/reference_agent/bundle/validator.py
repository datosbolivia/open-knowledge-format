# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import annotations

import dataclasses
import hashlib
import os
from pathlib import Path
import re
from typing import Any, Dict, List, Optional
import urllib.parse
import yaml

from reference_agent.bundle.document import OKFDocument

SHA256_PATTERN = re.compile(r"^sha256:[a-fA-F0-9]{64}$")
TRIAD_PATTERN = re.compile(r"^[a-zA-Z0-9_\-\.]+[:/][a-zA-Z0-9_\-\. ]+([:/][a-zA-Z0-9_\-\. ]+)?$")
VALID_POLICIES = {"allow_all", "zero_microdata", "deny"}
VALID_CONTRACT_TYPES = {"datapackage", "dcat", "openapi", "cwl", "cli_tool"}


@dataclasses.dataclass
class ValidationIssue:
    code: str
    severity: str  # ERROR | WARNING | INFO
    message: str
    path: Optional[str] = None
    line: Optional[int] = None

    def to_dict(self) -> Dict[str, Any]:
        return dataclasses.asdict(self)


@dataclasses.dataclass
class ValidationReport:
    valid: bool = True
    target: str = ""
    total_errors: int = 0
    total_warnings: int = 0
    issues: List[ValidationIssue] = dataclasses.field(default_factory=list)

    def add_issue(self, code: str, severity: str, message: str, path: Optional[str] = None, line: Optional[int] = None) -> None:
        self.issues.append(ValidationIssue(code=code, severity=severity, message=message, path=path, line=line))
        if severity == "ERROR":
            self.total_errors += 1
            self.valid = False
        elif severity == "WARNING":
            self.total_warnings += 1

    def to_dict(self) -> Dict[str, Any]:
        return {
            "valid": self.valid,
            "target": self.target,
            "total_errors": self.total_errors,
            "total_warnings": self.total_warnings,
            "issues": [i.to_dict() for i in self.issues],
        }


def validate_concept_content(content: str, file_path: str = "") -> ValidationReport:
    """Validates an individual OKF/ODKF markdown concept document."""
    report = ValidationReport(valid=True, target=file_path)

    lines = content.splitlines()
    if not lines or lines[0].strip() != "---":
        report.add_issue("PE1-SYNTAX", "ERROR", "Document must begin with YAML frontmatter delimiter '---'", path=file_path, line=1)
        return report

    end_idx = -1
    for i in range(1, len(lines)):
        if lines[i].strip() == "---":
            end_idx = i
            break

    if end_idx == -1:
        report.add_issue("PE1-SYNTAX", "ERROR", "Unclosed YAML frontmatter: missing ending delimiter '---'", path=file_path, line=len(lines))
        return report

    frontmatter_text = "\n".join(lines[1:end_idx])
    try:
        fm = yaml.safe_load(frontmatter_text)
        if not isinstance(fm, dict):
            report.add_issue("RULE-01-YAML", "ERROR", "Frontmatter must parse into a YAML mapping", path=file_path, line=1)
            return report
    except Exception as e:
        report.add_issue("RULE-01-YAML", "ERROR", f"YAML frontmatter parse error: {e}", path=file_path, line=1)
        return report

    # RULE-02 / PE2: Mandatory type
    concept_type = str(fm.get("type", "")).strip()
    if not concept_type:
        report.add_issue("PE2-TYPE-REQUIRED", "ERROR", "Field 'type' is required in concept frontmatter", path=file_path, line=1)

    norm_type = concept_type.lower()

    # RULE-03: Semantic Anchors
    skos = fm.get("skos", [])
    if norm_type == "dataset" and not skos:
        report.add_issue("RULE-03-SKOS-MISSING", "WARNING", "Dataset concept should define at least one ontological anchor in 'skos'", path=file_path, line=1)
    if isinstance(skos, list):
        for s in skos:
            parsed = urllib.parse.urlparse(str(s))
            if not parsed.scheme or not parsed.netloc:
                report.add_issue("RULE-03-SKOS-INVALID", "ERROR", f"Invalid URI in skos anchor: '{s}'", path=file_path, line=1)

    # RULE-04 / PE4: Syntactic Contracts
    contracts = fm.get("contracts", [])
    if norm_type == "dataset" and not contracts:
        report.add_issue("PE4-CONTRACT-MISSING", "WARNING", "Dataset concept defines no syntactic contract in 'contracts'", path=file_path, line=1)
    if isinstance(contracts, list):
        for idx, c in enumerate(contracts):
            if not isinstance(c, dict):
                report.add_issue("PE4-CONTRACT-INVALID", "ERROR", f"Contract[{idx}] must be a mapping", path=file_path, line=1)
                continue
            c_type = str(c.get("type", "")).strip().lower()
            c_path = str(c.get("path", "")).strip()
            if not c_type:
                report.add_issue("PE4-CONTRACT-TYPE", "ERROR", f"Contract[{idx}] is missing 'type'", path=file_path, line=1)
            elif c_type not in VALID_CONTRACT_TYPES:
                report.add_issue("PE4-CONTRACT-UNKNOWN", "WARNING", f"Contract[{idx}] has non-standard type '{c_type}'", path=file_path, line=1)

            if not c_path:
                report.add_issue("PE4-CONTRACT-PATH", "ERROR", f"Contract[{idx}] is missing 'path'", path=file_path, line=1)
            elif not c_path.startswith(("http://", "https://")) and file_path:
                base_dir = os.path.dirname(file_path)
                target_p = os.path.join(base_dir, c_path)
                if not os.path.exists(target_p):
                    report.add_issue("PE4-CONTRACT-LOCAL-NOT-FOUND", "WARNING", f"Local contract path '{c_path}' not found at '{target_p}'", path=file_path, line=1)

    # RULE-05 / PE5: Cryptographic Hash Integrity
    lineage = fm.get("lineage")
    if isinstance(lineage, dict):
        transformations = lineage.get("transformations", [])
        if isinstance(transformations, list):
            for idx, t in enumerate(transformations):
                if isinstance(t, dict):
                    h = str(t.get("hash", "")).strip()
                    tool = str(t.get("tool", "")).strip()
                    if h:
                        if not SHA256_PATTERN.match(h):
                            report.add_issue("PE5-HASH-FORMAT", "ERROR", f"Transformation[{idx}] hash '{h}' does not match format 'sha256:<hex64>'", path=file_path, line=1)
                        elif tool and file_path:
                            base_dir = os.path.dirname(file_path)
                            tool_file = os.path.join(base_dir, tool)
                            if os.path.isfile(tool_file):
                                with open(tool_file, "rb") as f:
                                    computed = hashlib.sha256(f.read()).hexdigest().lower()
                                expected = h.lower().replace("sha256:", "")
                                if computed != expected:
                                    report.add_issue("PE5-HASH-MISMATCH", "ERROR", f"Computed SHA-256 ('{computed}') of tool '{tool}' does not match declared hash ('{expected}')", path=file_path, line=1)

    # RULE-06: Policy Enum
    policy = fm.get("policy")
    if policy:
        p_str = str(policy).strip().lower()
        if p_str not in VALID_POLICIES:
            report.add_issue("RULE-06-POLICY-INVALID", "ERROR", f"Policy '{policy}' is invalid. Must be allow_all, zero_microdata, or deny", path=file_path, line=1)

    return report


def validate_triad(triad: str) -> ValidationReport:
    """Validates canonical triad addressing format."""
    report = ValidationReport(valid=True, target=triad)
    clean = triad.strip().strip("\"'")
    if not TRIAD_PATTERN.match(clean):
        report.add_issue("RULE-08-TRIAD-SYNTAX", "ERROR", f"Triad '{clean}' does not match canonical pattern 'catalog:dataset:resource' or 'catalog/dataset/resource'", path="")
    return report


def validate_bundle(bundle_dir: str) -> ValidationReport:
    """Recursively validates an entire OKF/ODKF bundle directory."""
    report = ValidationReport(valid=True, target=bundle_dir)
    target_path = Path(bundle_dir)

    if not target_path.exists() or not target_path.is_dir():
        report.add_issue("IO-DIR-NOT-FOUND", "ERROR", f"Bundle path '{bundle_dir}' is not an accessible directory", path=bundle_dir)
        return report

    concept_count = 0
    for root, _, files in os.walk(bundle_dir):
        for fname in files:
            if not fname.endswith(".md"):
                continue
            full_path = os.path.join(root, fname)

            if fname in ("index.md", "log.md"):
                continue

            try:
                with open(full_path, "r", encoding="utf-8") as f:
                    content = f.read()
                concept_count += 1
                doc_report = validate_concept_content(content, file_path=full_path)
                if not doc_report.valid:
                    report.valid = False
                report.total_errors += doc_report.total_errors
                report.total_warnings += doc_report.total_warnings
                report.issues.extend(doc_report.issues)
            except Exception as e:
                report.add_issue("IO-READ-ERROR", "ERROR", f"Failed to read '{full_path}': {e}", path=full_path)

    if concept_count == 0:
        report.add_issue("RULE-07-EMPTY-BUNDLE", "ERROR", "Bundle directory contains no concept markdown (.md) documents", path=bundle_dir)

    return report
