# `usecases/assets` — Embedded Cytoscape.js Bundle Viewer Assets (`src/usecases/assets`)

## Overview & Role

Static HTML, CSS, and JavaScript template files embedded at compile time into the `usecases` package (`//go:embed assets/*` in [`../visualize.go`](../visualize.go)) to produce self-contained interactive HTML visualizations of Open Knowledge Format (OKF v0.2) bundles.

## Interface

| File | Placeholder / Symbol (`[file:symbol]`) | Responsibility |
| :--- | :--- | :--- |
| [`viz.html`](./viz.html) | `[viz.html:/*__VIZ_CSS__*/]`, `[viz.html:/*__VIZ_JS__*/]`, `[viz.html:__BUNDLE_NAME__]`, `[viz.html:__BUNDLE_DATA__]` | HTML5 shell loading Cytoscape.js (`3.28.1`) and Marked (`12.0.0`) from jsDelivr CDN, declaring the graph canvas (`#graph`), search/filter/layout controls, and concept detail sidebar (`#detail`). |
| [`viz.css`](./viz.css) | Injected into `<style>` in [`viz.html`](./viz.html) | Stylesheet for the split-pane graph viewer, node type chips, trust/status/staleness badges, and rendered Markdown body. |
| [`viz.js`](./viz.js) | `[viz.js:clearSelection]`, `[viz.js:showDetail]`, `[viz.js:makeBadge]`, `[viz.js:formatActorEvent]`, `[viz.js:rewriteInternalLinks]` | Initializes the Cytoscape.js instance from `window.BUNDLE` and `window.BUNDLE_NAME`, populates `#filter-type`, computes reverse-link backlinks (`Cited by`), renders concept Markdown bodies via `marked.parse`, and rewrites internal `/<concept>.md` links to in-viewer node selection. |

## Invariants

1. **Zero Backend Dependency**: Generated `viz.html` artifacts inline all CSS, JS, and serialized `GraphPayload` JSON (`window.BUNDLE`) into a single portable file (`[../visualize.go:VisualizeUseCase.Generate]`).
2. **Visual Epistemic Signals (`[viz.js:showDetail]`, `[viz.js:makeBadge]`)**:
   - Deprecated nodes (`status = "deprecated"`) render with reduced opacity (`0.55`).
   - Stale nodes (`stale = true`) render with a dashed red border (`#b91c1c`).
   - Detail badges display `status`, derived `trust_tier` (`human-reviewed`, `machine-confirmed`, `unverified`), and `stale_after` freshness state.
3. **Internal Link Rewriting (`[viz.js:rewriteInternalLinks]`)**: Markdown links matching `/<concept>.md` where `<concept>` exists in `nodeIndex` are intercepted to pan/zoom and select the target node inside Cytoscape instead of navigating away.

## Side Effects

- **DOM & Browser State (`[viz.js:showDetail]`, `[viz.js:clearSelection]`)**: Mutates browser DOM nodes (`#graph`, `#detail`, `#filter-type`, `#search`, `#detail-type`, `#detail-badges`) and initializes the client-side `cytoscape` instance when loaded in a browser.

## Verification

```bash
cd src && go test -v -run TestVisualize ./usecases/...
```
