import os
import pytest
from reference_agent.bundle.validator import (
    validate_concept_content,
    validate_triad,
    validate_bundle,
)

def test_validate_valid_concept():
    content = """---
type: dataset
title: Atlas Electoral
description: Resultados oficiales
skos:
  - https://www.wikidata.org/entity/Q39825
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
policy: allow_all
---

# Concept Body
"""
    rep = validate_concept_content(content)
    assert rep.valid is True
    assert rep.total_errors == 0

def test_validate_missing_type():
    content = """---
title: Missing Type
---
Body
"""
    rep = validate_concept_content(content)
    assert rep.valid is False
    assert any(i.code == "PE2-TYPE-REQUIRED" for i in rep.issues)

def test_validate_invalid_skos():
    content = """---
type: dataset
skos:
  - not-a-valid-uri
contracts:
  - type: datapackage
    path: https://example.org/dp.json
---
Body
"""
    rep = validate_concept_content(content)
    assert rep.valid is False
    assert any(i.code == "RULE-03-SKOS-INVALID" for i in rep.issues)

def test_validate_triad():
    assert validate_triad("bolivia:elecciones:votos").valid is True
    assert validate_triad("bolivia/elecciones/votos").valid is True
    assert validate_triad("invalid_format").valid is False

def test_validate_bundle(tmp_path):
    # Empty dir gives error
    rep = validate_bundle(str(tmp_path))
    assert rep.valid is False

    # Add concept
    doc_path = tmp_path / "dataset.md"
    doc_path.write_text("""---
type: dataset
title: Valid
skos:
  - https://schema.org/Dataset
contracts:
  - type: datapackage
    path: https://example.com/dp.json
---
Body
""", encoding="utf-8")

    rep = validate_bundle(str(tmp_path))
    assert rep.valid is True
