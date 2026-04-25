# Virtual Microscope — Technical Design

> This document contains technology-specific decisions, tooling choices, and implementation guidance for the Virtual Microscope Viewer feature.
> For the abstract, technology-agnostic requirements see [Virtual Microscope Viewer in REQUIREMENTS.md](REQUIREMENTS.md#virtual-microscope-viewer).
> For term definitions see [Requirements Glossary](REQUIREMENTS-GLOSSARY.md).

---

## Architecture: S3 Direct Fetch Model

This is the most important architectural decision for the viewer. The app does **not** proxy image tiles through the Go backend. Instead it uses a decoupled model where the browser fetches tiles directly from S3.

```
Student's Browser
      │
      │  1. GET /slides/42          (one DB query)
      ▼
  Go Backend  ──────────────────►  PostgreSQL
      │
      │  2. Returns HTML page + dzi_url
      ▼
Student's Browser
      │
      │  3. GET <dzi_url>           (tiny .dzi text file from S3)
      ▼
    AWS S3
      │
      │  4. OpenSeadragon calculates which ~15 tiles are visible
      │     and fires parallel requests directly to S3
      ▼
    AWS S3  ──► 256×256 tile ──► Browser (repeated per pan/zoom)
```

**Why this matters:**
- The Go backend handles exactly **one** database query per slide page load, regardless of how many users are zooming simultaneously.
- If 50 students are all panning at once, S3 absorbs ~5,000 tile requests. Aurora/PostgreSQL sees nothing.
- Your Go app stays free to handle business logic; it never touches image bytes at runtime.

**What the Go backend is responsible for:**
- Serving the initial HTML page with the `dzi_url` embedded.
- The `GET /api/slides/:id/dzi` metadata endpoint (see [Viewer Metadata API](#viewer-metadata-api-fr-vm-3)).
- The tiling pipeline CLI (offline, not in the request path).

**What the browser (OpenSeadragon) is responsible for:**
- Parsing the `.dzi` descriptor to understand the image pyramid dimensions.
- Calculating which tiles are needed for the current viewport on every pan/zoom event.
- Fetching those tiles in parallel directly from S3.
- Compositing tiles into the visible canvas.

---

## The Image Pyramid

A DZI is not a single image — it is a **pyramid** of the same image at decreasing resolutions.

```
Level 0  (thumbnail, ~1×1 tile)
Level 1  (~2×2 tiles)
...
Level N  (full resolution, e.g. 200×150 tiles for a 40× slide)
```

- When the student is zoomed out, OpenSeadragon fetches tiles from a low level (few tiles, fast).
- When zoomed in to 40×, it fetches tiles from the highest level (only the ~15 tiles currently on screen).
- The student never downloads the full image — only the visible slice at the appropriate resolution.

`vips dzsave` generates all pyramid levels automatically. The `--depth onetile` flag controls how deep the pyramid goes.

---

## CORS Configuration (AWS S3)

Because the browser fetches tiles from `s3.amazonaws.com` while the page is served from `tissquest.com`, S3 will block the requests by default (same-origin policy). You must configure a CORS policy on the S3 bucket.

**Minimum required S3 CORS policy** (JSON, set via AWS Console or CLI):

```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["GET"],
    "AllowedOrigins": ["https://tissquest.com"],
    "ExposeHeaders": [],
    "MaxAgeSeconds": 3600
  }
]
```

For local development, add `http://localhost:8080` (or your dev port) to `AllowedOrigins`.

**Apply via AWS CLI:**
```bash
aws s3api put-bucket-cors \
  --bucket tissquest \
  --cors-configuration file://cors.json
```

> Without this step, OpenSeadragon will silently fail to load any tiles. The viewer will appear blank. This is the most common first-time setup mistake.

---

## Image Tiling (FR-VM-1)

**Format**: [Deep Zoom Image (DZI)](https://docs.microsoft.com/en-us/previous-versions/windows/silverlight/dotnet-windows-silverlight/cc645050(v=vs.95)) — one `.dzi` XML descriptor + a companion `_files/` directory of 256×256 JPEG tiles.

**Tool**: [`libvips`](https://www.libvips.org/) via the `vips dzsave` command.

```bash
vips dzsave source.tiff output --tile-size 256 --overlap 1 --depth onetile --suffix .jpg[Q=85]
```

**Storage**: AWS S3. Tile sets are uploaded under the path `slides/{slide_id}/` immediately after processing. The `.dzi` file is served at `slides/{slide_id}/output.dzi`.

**Pipeline trigger**: A Go CLI command (`cmd/tile-pipeline`) that accepts a slide ID, downloads the source image, runs `vips dzsave`, and uploads the result to S3, then patches the Slide record with the resulting `dzi_url`.

---

## Slide Data Model (FR-VM-2)

New columns on the `slides` table (GORM migration):

| Column | Type | Notes |
|---|---|---|
| `dzi_url` | `TEXT` | Full S3 URL to the `.dzi` descriptor. Nullable. |
| `base_magnification` | `INTEGER` | Objective used at capture (4, 10, 40, 100). |
| `microns_per_pixel` | `FLOAT` | Spatial calibration. Derived from objective + camera sensor spec. |
| `home_viewport` | `JSONB` | OpenSeadragon `{x, y, zoom}` viewport object. Nullable. |

---

## Viewer Metadata API (FR-VM-3)

**Endpoint**: `GET /api/slides/:id/dzi`

**Response**:
```json
{
  "dzi_url": "https://s3.amazonaws.com/tissquest/slides/42/output.dzi",
  "base_magnification": 40,
  "microns_per_pixel": 0.2505,
  "tile_size": 256,
  "home_viewport": { "x": 0.5, "y": 0.5, "zoom": 4.2 }
}
```

Handler lives in `cmd/api-server-gin/slides/` and delegates to `SlideService`. Returns `404` if the slide has no `dzi_url`.

**Home view update**: `PATCH /api/slides/:id/home-viewport` — body `{"x": float, "y": float, "zoom": float}`.

---

## Frontend: OpenSeadragon (FR-VM-4)

**Library**: [OpenSeadragon v4.x](https://openseadragon.github.io/) loaded from CDN or bundled locally.

**Recommended initialization**:
```javascript
const viewer = OpenSeadragon({
  id: "microscope-viewer",
  prefixUrl: "/static/openseadragon/images/",
  tileSources: dziMetadata.dzi_url,
  showNavigationControl: true,
  zoomPerScroll: 2,        // coarse focus knob feel
  animationTime: 0.5,      // mechanical glide
  blendingTime: 0.1,
  constrainDuringPan: true,
  maxZoomPixelRatio: 2,
  visibilityRatio: 1.0,
});
```

The viewer container `#microscope-viewer` is set to `height: calc(100vh - <header height>)` via CSS.

---

## Objective Lens Switcher (FR-VM-5)

Zoom levels are computed relative to `base_magnification`:

```javascript
const objectives = [4, 10, 40];

function zoomToObjective(target) {
  const ratio = target / dziMetadata.base_magnification;
  viewer.viewport.zoomTo(ratio);
}
```

The switcher is an HTML `<div role="group" aria-label="Objective lens">` with `<button>` elements. Active state is toggled via a CSS class on `viewer.addHandler('zoom', ...)`.

---

## Scale Bar (FR-VM-6)

**Plugin**: [OpenSeadragon Scalebar](https://github.com/usnistgov/OpenSeadragonScalebar)

```javascript
viewer.scalebar({
  type: OpenSeadragon.ScalebarType.MICROSCOPY,
  pixelsPerMeter: (1 / dziMetadata.microns_per_pixel) * 1e6,
  minWidth: "75px",
  location: OpenSeadragon.ScalebarLocation.BOTTOM_LEFT,
  xOffset: 10,
  yOffset: 10,
  stayInsideImage: false,
  color: "white",
  fontColor: "white",
  backgroundColor: "rgba(0,0,0,0.5)",
  barThickness: 3,
});
```

---

## Viewport Coordinates

OpenSeadragon does not use pixel coordinates. It uses a **normalized coordinate system** where the full image width equals `1.0`.

- Top-left corner of the image = `(0.0, 0.0)`
- Center of the image = `(0.5, aspect_ratio / 2)`
- A point at 23% from the left, 88% from the top = `(0.23, 0.88)`

This means viewport coordinates are **resolution-independent** — the same `{x, y, zoom}` values work correctly regardless of the source image dimensions. This is why `home_viewport` stored in the database is stable even if the source image is re-tiled at a different resolution.

**Zoom** is expressed as a multiplier relative to the "fit to screen" state. `zoom: 1.0` means the full image fits the viewer. `zoom: 4.5` means the image is 4.5× larger than fit-to-screen.

---

## Home View (FR-VM-7)

On viewer `open` event, if `home_viewport` is present in the metadata response:
```javascript
viewer.addHandler('open', () => {
  if (dziMetadata.home_viewport) {
    const { x, y, zoom } = dziMetadata.home_viewport;
    viewer.viewport.panTo(new OpenSeadragon.Point(x, y), true);
    viewer.viewport.zoomTo(zoom, null, true);
  }
});
```

The "Set Home" button is rendered only when the page is loaded in edit mode (Go template conditional). It calls `PATCH /api/slides/:id/home-viewport` with the current `viewer.viewport` state serialized to JSON.

---

## Mobile Support (FR-VM-8)

OpenSeadragon handles touch natively. Additional considerations:
- The objective switcher uses `min-height: 44px` touch targets.
- The viewer container uses `touch-action: none` to prevent browser scroll interference.
- Tested on iOS Safari 17+ and Android Chrome 120+.

---

## Static Fallback (FR-VM-9)

The Slide detail Go template checks for `dzi_url`:
```html
{{ if .Slide.DziURL }}
  <div id="microscope-viewer"></div>
  <!-- OpenSeadragon init script -->
{{ else }}
  <img src="{{ .Slide.ImageURL }}" alt="{{ .Slide.Description }}">
{{ end }}
```

No JavaScript is loaded for the fallback path.

---

## Implementation Plan

Ordered for maximum testability — each step leaves the system in a working state.

| Step | Scope | Files touched | Deliverable |
|---|---|---|---|
| **1** | Domain + data model | `slide.go`, `slide_model.go`, `gorm_slide_repository.go` | Existing slides unaffected; new fields nullable/zero |
| **2** | Tiling pipeline CLI | `cmd/tile-pipeline/main.go` (new) | Run manually per slide; produces real DZI tiles in S3 |
| **3** | Metadata API | `slides/slides.go`, `repository_interface.go`, `slide_service.go` | `GET /api/slides/:id/dzi` + `PATCH /api/slides/:id/home-viewport` |
| **4** | Viewer page | Slide detail template + JS | OpenSeadragon viewer with objective switcher and scale bar |

**Deferred to post-MVP:**
- Automatic pipeline trigger on image upload
- "Set Home" button in the UI
- Mobile polish
- A new service method for the DZI endpoint (`GetByID` is sufficient once Step 1 is done)

---

## Timeline

| Month | Milestone |
|---|---|
| May 2026 | Tiling pipeline working; 20 slides processed and stored in S3 |
| June 2026 | OpenSeadragon integrated; objective switcher and scale bar functional |
| July 2026 | Home view, mobile testing, Docker deployment |

---

## References
- [OpenSeadragon Documentation](https://openseadragon.github.io/docs/)
- [OpenSeadragon Scalebar Plugin](https://github.com/usnistgov/OpenSeadragonScalebar)
- [libvips dzsave documentation](https://www.libvips.org/API/current/VipsForeignSave.html#vips-dzsave)
- [Deep Zoom Image format spec](https://docs.microsoft.com/en-us/previous-versions/windows/silverlight/dotnet-windows-silverlight/cc645050(v=vs.95))
- [Requirements Specification](REQUIREMENTS.md)
- [Requirements Glossary](REQUIREMENTS-GLOSSARY.md)
