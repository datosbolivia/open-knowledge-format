package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"gopkg.in/yaml.v3"
)

// SyncDataPackageDistributions safely synchronizes automated physical Distribution fields
// (path, hash_sha256, bytes_count, format, updated_at, quality_profile) and appends newly
// discovered schema fields into datapackage.yaml without removing or overwriting human-curated
// knowledge (descriptions, concept anchors, value_labels, or existing fields).
// Discrepancies against curated columns emit DriftAlerts instead of mutating curated definitions.
func SyncDataPackageDistributions(bundleDir string, outputs []domain.ProvenanceEvent) ([]domain.DriftAlert, []string, error) {
	if len(outputs) == 0 || strings.TrimSpace(bundleDir) == "" {
		return nil, nil, nil
	}

	pkgPath := filepath.Join(bundleDir, "datapackage.yaml")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		alt := filepath.Join(bundleDir, "datapackage.yml")
		if _, err2 := os.Stat(alt); err2 == nil {
			pkgPath = alt
		}
	}

	pkgMap := make(map[string]interface{})
	if raw, err := os.ReadFile(pkgPath); err == nil {
		if err := yaml.Unmarshal(raw, &pkgMap); err != nil {
			return nil, nil, fmt.Errorf("failed parsing %s: %w", pkgPath, err)
		}
	} else if os.IsNotExist(err) {
		pkgMap["name"] = domain.Slugify(filepath.Base(bundleDir))
		pkgMap["profile"] = "tabular-data-package"
	} else {
		return nil, nil, err
	}

	var resources []map[string]interface{}
	if rawRes, ok := pkgMap["resources"].([]interface{}); ok {
		for _, item := range rawRes {
			if m, ok := item.(map[string]interface{}); ok {
				resources = append(resources, m)
			}
		}
	}

	var drifts []domain.DriftAlert
	var synced []string
	nowISO := time.Now().UTC().Format(time.RFC3339)

	for _, out := range outputs {
		resName := strings.TrimSpace(out.Resource)
		resPath := strings.TrimSpace(out.Path)
		if resName == "" && resPath != "" {
			base := filepath.Base(resPath)
			resName = strings.TrimSuffix(base, filepath.Ext(base))
		}
		if resName == "" {
			continue
		}

		if resPath != "" && filepath.IsAbs(resPath) {
			if rel, err := filepath.Rel(bundleDir, resPath); err == nil && !strings.HasPrefix(rel, "..") {
				resPath = filepath.ToSlash(rel)
			}
		}

		idx := -1
		for i, r := range resources {
			existingName, _ := r["name"].(string)
			existingPath, _ := r["path"].(string)
			if strings.EqualFold(existingName, resName) || (resPath != "" && existingPath == resPath) {
				idx = i
				break
			}
		}

		if idx == -1 {
			newRes := map[string]interface{}{
				"name":       resName,
				"updated_at": nowISO,
			}
			if resPath != "" {
				newRes["path"] = resPath
			}
			if out.Format != "" {
				newRes["format"] = out.Format
			}
			if out.HashSHA256 != "" {
				newRes["hash_sha256"] = out.HashSHA256
			}
			if out.BytesCount > 0 {
				newRes["bytes_count"] = out.BytesCount
			}
			if out.RowCount > 0 || out.Completeness > 0 {
				qp := map[string]interface{}{}
				if out.RowCount > 0 {
					qp["row_count"] = out.RowCount
				}
				if out.Completeness > 0 {
					qp["completeness"] = out.Completeness
				}
				newRes["quality_profile"] = qp
			}
			if len(out.Fields) > 0 {
				fieldsList := make([]map[string]interface{}, 0, len(out.Fields))
				for _, f := range out.Fields {
					fm := map[string]interface{}{"name": f.Name}
					if f.Type != "" {
						fm["type"] = f.Type
					}
					if f.Description != "" {
						fm["description"] = f.Description
					}
					fieldsList = append(fieldsList, fm)
				}
				newRes["schema"] = map[string]interface{}{"fields": fieldsList}
			}
			resources = append(resources, newRes)
			synced = append(synced, resName)
			continue
		}

		// Existing resource: update physical distribution attributes only; preserve curated schema & metadata
		target := resources[idx]
		if resPath != "" {
			target["path"] = resPath
		}
		if out.Format != "" {
			target["format"] = out.Format
		}
		if out.HashSHA256 != "" {
			target["hash_sha256"] = out.HashSHA256
		}
		if out.BytesCount > 0 {
			target["bytes_count"] = out.BytesCount
		}
		target["updated_at"] = nowISO

		if out.RowCount > 0 || out.Completeness > 0 {
			qp, _ := target["quality_profile"].(map[string]interface{})
			if qp == nil {
				qp = map[string]interface{}{}
			}
			if out.RowCount > 0 {
				qp["row_count"] = out.RowCount
			}
			if out.Completeness > 0 {
				qp["completeness"] = out.Completeness
			}
			target["quality_profile"] = qp
		}

		if len(out.Fields) > 0 {
			schemaMap, _ := target["schema"].(map[string]interface{})
			if schemaMap == nil {
				schemaMap = map[string]interface{}{}
			}
			var existingFields []map[string]interface{}
			if rawF, ok := schemaMap["fields"].([]interface{}); ok {
				for _, item := range rawF {
					if m, ok := item.(map[string]interface{}); ok {
						existingFields = append(existingFields, m)
					}
				}
			}

			observedByName := make(map[string]domain.ResourceField, len(out.Fields))
			for _, f := range out.Fields {
				observedByName[strings.ToLower(strings.TrimSpace(f.Name))] = f
			}

			existingByName := make(map[string]bool, len(existingFields))
			for _, ef := range existingFields {
				efName, _ := ef["name"].(string)
				key := strings.ToLower(strings.TrimSpace(efName))
				if key == "" {
					continue
				}
				existingByName[key] = true
				obs, found := observedByName[key]
				if !found {
					drifts = append(drifts, domain.DriftAlert{
						Resource: resName,
						Kind:     "missing_column",
						Field:    efName,
						Expected: efName,
						Actual:   "<absent>",
						Severity: "WARNING",
						Message:  fmt.Sprintf("Curated field '%s' in resource '%s' was not present in ETL output", efName, resName),
					})
					continue
				}
				efType, _ := ef["type"].(string)
				if efType != "" && obs.Type != "" && !strings.EqualFold(efType, obs.Type) {
					drifts = append(drifts, domain.DriftAlert{
						Resource: resName,
						Kind:     "type_changed",
						Field:    efName,
						Expected: efType,
						Actual:   obs.Type,
						Severity: "WARNING",
						Message:  fmt.Sprintf("Field '%s' in resource '%s' changed type from '%s' to '%s' (curated type preserved)", efName, resName, efType, obs.Type),
					})
				}
			}

			// Additive-only: append new fields discovered in output
			for _, obs := range out.Fields {
				key := strings.ToLower(strings.TrimSpace(obs.Name))
				if key != "" && !existingByName[key] {
					nf := map[string]interface{}{"name": obs.Name}
					if obs.Type != "" {
						nf["type"] = obs.Type
					}
					if obs.Description != "" {
						nf["description"] = obs.Description
					}
					existingFields = append(existingFields, nf)
				}
			}
			schemaMap["fields"] = existingFields
			target["schema"] = schemaMap
		}

		synced = append(synced, resName)
	}

	pkgMap["resources"] = resources
	encoded, err := yaml.Marshal(pkgMap)
	if err != nil {
		return drifts, synced, fmt.Errorf("failed marshaling %s: %w", pkgPath, err)
	}
	if err := os.WriteFile(pkgPath, encoded, 0644); err != nil {
		return drifts, synced, fmt.Errorf("failed writing %s: %w", pkgPath, err)
	}

	return drifts, synced, nil
}
