#!/usr/bin/env python3
"""
read_ckan.py — Lector y Cosechador de Instancias CKAN a Paquetes OKF v0.2.

Permite consultar, filtrar y exportar paquetes desde cualquier portal CKAN (Action API v3)
hacia la estructura estándar de Open Knowledge Format (index.md + datapackage.yaml).

Compatible con 'okf run' para seguimiento de procedencia y linaje criptográfico.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import unicodedata
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

try:
    import yaml
except ImportError:
    yaml = None


def slugify(text: str) -> str:
    """Normaliza un texto a slug alfanumérico ASCII."""
    norm = unicodedata.normalize("NFKD", text).encode("ASCII", "ignore").decode("utf-8")
    cleaned = re.sub(r"[^a-zA-Z0-9]+", "_", norm.strip().lower())
    return cleaned.strip("_") or "dataset"


def fetch_json(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 15) -> Dict[str, Any]:
    """Realiza una petición HTTP GET segura a un endpoint JSON."""
    parsed = urllib.parse.urlparse(url)
    if parsed.scheme not in ("http", "https"):
        raise ValueError(f"Esquema no permitido: {parsed.scheme}")

    req_headers = {
        "Accept": "application/json",
        "User-Agent": "okf-ckan-reader/0.2.2 (OpenKnowledgeFormat)",
    }
    if headers:
        req_headers.update(headers)

    req = urllib.request.Request(url, headers=req_headers)
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        if resp.status != 200:
            raise RuntimeError(f"HTTP {resp.status} al consultar {url}")
        body = resp.read().decode("utf-8")
        return json.loads(body)


def ckan_package_to_okf(pkg: Dict[str, Any], base_url: str) -> Tuple[Dict[str, Any], str]:
    """
    Transforma un paquete crudo de CKAN Action API v3 en:
    1. Manifiesto Frictionless DataPackage (dict para datapackage.yaml)
    2. Documento de concepto OKF v0.2 (texto Markdown para index.md)
    """
    pkg_id = pkg.get("id", "")
    pkg_name = slugify(pkg.get("name") or pkg.get("title") or "ckan_package")
    title = (pkg.get("title") or pkg.get("name") or "Dataset CKAN").strip()
    notes = (pkg.get("notes") or "Sin descripción provista por el portal CKAN.").strip()

    # Extraer extras como metadatos extendidos
    extras_dict: Dict[str, Any] = {}
    for extra in pkg.get("extras", []):
        if isinstance(extra, dict) and "key" in extra and "value" in extra:
            k = extra["key"]
            v = extra["value"]
            try:
                extras_dict[k] = json.loads(v)
            except Exception:
                extras_dict[k] = v

    # Procesar recursos tabulares / distribuciones
    resources: List[Dict[str, Any]] = []
    for idx, res in enumerate(pkg.get("resources", [])):
        r_name = slugify(res.get("name") or f"recurso_{idx + 1}")
        r_format = (res.get("format") or "csv").strip().lower()
        r_url = res.get("url") or ""
        r_desc = res.get("description") or f"Recurso {r_name} del dataset {title}"

        res_entry: Dict[str, Any] = {
            "name": r_name,
            "path": r_url,
            "format": r_format,
            "description": r_desc,
        }
        if res.get("size"):
            res_entry["bytes"] = res["size"]
        if res.get("mimetype"):
            res_entry["mediatype"] = res["mimetype"]

        resources.append(res_entry)

    # 1. Manifiesto Frictionless DataPackage
    manifest: Dict[str, Any] = {
        "name": pkg_name,
        "title": title,
        "description": notes,
        "version": pkg.get("version") or "1.0.0",
        "resources": resources,
    }

    if pkg.get("organization") and isinstance(pkg["organization"], dict):
        manifest["sources"] = [{
            "name": pkg["organization"].get("title") or pkg["organization"].get("name"),
            "path": f"{base_url.rstrip('/')}/organization/{pkg['organization'].get('name', '')}",
        }]

    # 2. Documento OKF index.md
    frontmatter_dict: Dict[str, Any] = {
        "type": "dataset",
        "id": pkg_name,
        "title": title,
        "description": notes.split("\n")[0][:160],
        "contracts": [
            {
                "type": "datapackage",
                "path": "./datapackage.yaml" if yaml else "./datapackage.json",
            }
        ],
        "sources": [
            {
                "name": f"Portal CKAN ({base_url})",
                "url": f"{base_url.rstrip('/')}/dataset/{pkg.get('name', pkg_id)}",
            }
        ],
    }

    # Map skos tags or default dataset concept anchor
    skos_anchors: List[str] = []
    if "skos" in extras_dict:
        if isinstance(extras_dict["skos"], list):
            skos_anchors.extend(extras_dict["skos"])
        elif isinstance(extras_dict["skos"], str):
            skos_anchors.append(extras_dict["skos"])
    for tag in pkg.get("tags", []):
        if isinstance(tag, dict) and "name" in tag:
            tag_name = tag["name"]
            if tag_name.startswith("http://") or tag_name.startswith("https://"):
                skos_anchors.append(tag_name)
    if not skos_anchors:
        # Default anchor: Wikidata Q1172284 (open data)
        skos_anchors.append("https://www.wikidata.org/wiki/Q1172284")
    frontmatter_dict["skos"] = skos_anchors

    if "spatial" in extras_dict:
        frontmatter_dict["spatial"] = extras_dict["spatial"]
    if "temporal" in extras_dict:
        frontmatter_dict["temporal"] = extras_dict["temporal"]

    # Renderizar Markdown
    fm_lines = ["---"]
    for k, v in frontmatter_dict.items():
        if isinstance(v, (dict, list)):
            if yaml:
                fm_lines.append(f"{k}:")
                for sub in yaml.dump(v, allow_unicode=True).strip().split("\n"):
                    fm_lines.append(f"  {sub}")
            else:
                fm_lines.append(f"{k}: {json.dumps(v, ensure_ascii=False)}")
        else:
            fm_lines.append(f"{k}: {v}")
    fm_lines.append("---\n")

    md_body = [
        f"# {title}\n",
        f"{notes}\n",
        "## Recursos del Dataset\n",
    ]
    for r in resources:
        md_body.append(f"- **`{r['name']}`** (`{r['format']}`): [{r['path']}]({r['path']}) — {r.get('description', '')}")

    md_body.append("\n## Procedencia y Cosecha\n")
    md_body.append(f"- **Origen CKAN**: [{base_url}]({base_url})")
    md_body.append(f"- **ID CKAN**: `{pkg_id}`")
    md_body.append("- **Cosechado por**: `okf run connectors/ckan/read_ckan.py`\n")

    index_md = "\n".join(fm_lines) + "\n".join(md_body)
    return manifest, index_md


def harvest_ckan(
    base_url: str,
    query: str = "",
    package_id: str = "",
    limit: int = 10,
    token: str = "",
    out_dir: Optional[str] = None,
) -> List[Dict[str, Any]]:
    """Consulta una instancia CKAN y opcionalmente exporta los bundles OKF resultantes."""
    clean_base = base_url.rstrip("/")
    headers = {}
    if token:
        headers["Authorization"] = token
        headers["X-CKAN-API-Key"] = token

    packages: List[Dict[str, Any]] = []

    if package_id:
        endpoint = f"{clean_base}/api/3/action/package_show?id={urllib.parse.quote(package_id)}"
        resp = fetch_json(endpoint, headers=headers)
        if resp.get("success") and "result" in resp:
            packages.append(resp["result"])
    else:
        q_param = f"&q={urllib.parse.quote(query)}" if query else ""
        endpoint = f"{clean_base}/api/3/action/package_search?rows={limit}{q_param}"
        resp = fetch_json(endpoint, headers=headers)
        if resp.get("success") and "result" in resp:
            packages = resp["result"].get("results", [])

    results: List[Dict[str, Any]] = []
    target_path = Path(out_dir).resolve() if out_dir else None
    if target_path:
        target_path.mkdir(parents=True, exist_ok=True)

    for pkg in packages:
        manifest, index_md = ckan_package_to_okf(pkg, clean_base)
        pkg_slug = manifest["name"]
        results.append({
            "name": pkg_slug,
            "title": manifest["title"],
            "resources_count": len(manifest["resources"]),
            "ckan_id": pkg.get("id"),
        })

        if target_path:
            pkg_dir = target_path / pkg_slug
            pkg_dir.mkdir(parents=True, exist_ok=True)

            # Escribir datapackage
            if yaml:
                dp_file = pkg_dir / "datapackage.yaml"
                with open(dp_file, "w", encoding="utf-8") as f:
                    yaml.dump(manifest, f, allow_unicode=True, sort_keys=False)
            else:
                dp_file = pkg_dir / "datapackage.json"
                with open(dp_file, "w", encoding="utf-8") as f:
                    json.dump(manifest, f, indent=2, ensure_ascii=False)

            # Escribir index.md
            with open(pkg_dir / "index.md", "w", encoding="utf-8") as f:
                f.write(index_md)

    return results


def main():
    parser = argparse.ArgumentParser(
        description="Lector y Cosechador de Instancias CKAN a Paquetes OKF v0.2",
    )
    parser.add_argument("--url", "-u", required=True, help="URL base de la instancia CKAN (ej. https://datos.gob.bo)")
    parser.add_argument("--query", "-q", default="", help="Filtro de búsqueda de datasets")
    parser.add_argument("--package-id", "-p", default="", help="Identificador o slug de un dataset específico")
    parser.add_argument("--limit", "-l", type=int, default=10, help="Límite de datasets a consultar")
    parser.add_argument("--token", "-t", default="", help="Token o API Key de autenticación")
    parser.add_argument("--out", "-o", default="", help="Directorio destino para guardar los bundles OKF")

    args = parser.parse_args()

    try:
        results = harvest_ckan(
            base_url=args.url,
            query=args.query,
            package_id=args.package_id,
            limit=args.limit,
            token=args.token,
            out_dir=args.out,
        )
        report = {
            "status": "success",
            "ckan_url": args.url,
            "total_packages": len(results),
            "packages": results,
            "output_directory": args.out or "(stdout)",
        }
        print(json.dumps(report, indent=2, ensure_ascii=False))
    except Exception as e:
        err_report = {
            "status": "error",
            "error": str(e),
            "ckan_url": args.url,
        }
        print(json.dumps(err_report, indent=2, ensure_ascii=False), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
