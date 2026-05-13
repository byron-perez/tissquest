# How to Add a Tissue Record with a Tiled Slide

This guide covers the complete workflow from a raw microscopy image on your local machine to a fully interactive virtual microscope slide visible in TissExplorer.

---

## Prerequisites

These need to be in place once. Skip if already done.

- **libvips installed** on your local machine:
  ```bash
  # Arch Linux
  sudo pacman -S libvips

  # Ubuntu / Debian
  sudo apt install libvips-tools

  # macOS
  brew install vips
  ```
- **`.env` file** present in the project root with valid AWS credentials and DB connection (see `.env.example`).
- **S3 bucket** (`tissdb`) with the public read policy applied to `slides/*` and CORS configured for your domain. See [Virtual Microscope — Technical Design](../VIRTUAL-MICROSCOPE-TECH.md#cors-configuration-aws-s3).

---

## What you need before starting

| Item | Example |
|---|---|
| A microscopy image file | `pteridium_frond_40x.jpg` |
| The objective used at capture | `40` |
| Microns per pixel (µm/px) for that objective | `0.2505` |

> **Tip on µm/px**: For a 40× objective on a typical lab camera, `0.25` is a safe default. For 10× use `1.0`, for 4× use `2.5`. You can refine this later with a stage micrometer image.

---

## Step 1 — Create the Tissue Record

1. Open the web interface and go to **Tejidos** (or navigate to `/tissue_records`).
2. Click **New Tissue Record** and fill in:
   - **Name** — e.g. "Fronda de helecho — corte transversal"
   - **Notes** — preparation details, anatomical context
   - **Taxon** — select or create the appropriate taxon
3. Save. Note the tissue record ID from the URL: `/tissue_records/7/workspace` → ID is `7`.

---

## Step 2 — Create the Slide

1. From the tissue record workspace, click **+ Add Slide**.
2. Fill in:
   - **Name** — e.g. "Corte transversal 40×"
   - **Magnification** — `40` (must match what you captured)
   - **Staining** — e.g. `H&E`, `Azul de metileno`
3. Save. Note the slide ID from the gallery card — you'll need it for the pipeline.

   > The slide ID is visible in the browser network tab when you interact with the card, or you can query the DB: `SELECT id, name FROM slides ORDER BY id DESC LIMIT 5;`

---

## Step 3 — Upload the source image

From the slide card in the workspace, click **Upload image** and select your microscopy image file. This uploads it to S3 at `slides/original/<slide_id>.png` and registers the `original` variant in the database.

After upload the card will show the image thumbnail.

---

## Step 4 — Run the tiling pipeline

From the project root on your local machine:

```bash
# Single slide — use this when you know the exact calibration
go run ./cmd/tile-pipeline \
  -slide <slide_id> \
  -magnification 40 \
  -microns-per-pixel 0.2505

# Single slide with a local file (skips the S3 download)
go run ./cmd/tile-pipeline \
  -slide <slide_id> \
  -magnification 40 \
  -microns-per-pixel 0.2505 \
  -source /path/to/pteridium_frond_40x.jpg

# Batch — tiles ALL slides that have an image but no DZI yet
go run ./cmd/tile-pipeline -batch

# Batch with explicit calibration fallback
go run ./cmd/tile-pipeline -batch -magnification 40 -microns-per-pixel 0.2505
```

**What the pipeline does:**
1. Downloads the source image from S3 (or uses `-source` directly)
2. Runs `vips dzsave` locally — generates a pyramid of 256×256 JPEG tiles
3. Uploads the `.dzi` descriptor and all tile folders to `s3://tissdb/slides/<id>/dzi/`
4. Updates the slide record in the database with `dzi_url`, `base_magnification`, `microns_per_pixel`

**Expected output:**
```
Connected to PostgreSQL database
Database migration completed successfully
downloading source image from s3://tissdb/slides/original/7.png
running: /usr/bin/vips dzsave /tmp/tissquest-src-xxx.png /tmp/tissquest-tile-7-xxx/output ...
uploading .dzi to s3://tissdb/slides/7/dzi/output.dzi
uploading tiles from /tmp/.../output_files to s3://tissdb/slides/7/dzi/output_files/
✓ slide 7 tiled
  dzi_url: https://tissdb.s3.sa-east-1.amazonaws.com/slides/7/dzi/output.dzi
```

> The two `VIPS-WARNING` lines about `vips-openslide.so` and `vips-magick.so` are harmless — those are optional modules not needed for JPEG/PNG tiling.

---

## Step 5 — Verify

1. Go back to the tissue record workspace. The slide card should now show a **🔬 View** button instead of **🧩 Tile**.
2. Click **🔬 View** — the virtual microscope should open with the slide loaded, objective buttons active, and scale bar visible.
3. The slide will also appear in **TissExplorer** (`/tissue_records`) with its thumbnail and a direct "🔬 Ver" link.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `vips: executable file not found` | libvips not installed | Run the install command for your OS above |
| `GetObject ... AccessDenied` | Source image not uploaded yet | Complete Step 3 first |
| `🔬 View` button doesn't appear after pipeline | DB not updated | Check pipeline output for `FAILED` lines; re-run |
| Viewer shows blank black screen | CORS not configured on S3 | See [CORS setup](../VIRTUAL-MICROSCOPE-TECH.md#cors-configuration-aws-s3) |
| Scale bar shows `NaN m` | `microns_per_pixel` is 0 in DB | Re-run pipeline with explicit `-microns-per-pixel` flag |

---

## Quick reference — full workflow in one block

```bash
# 1. Create tissue record + slide via web UI, note slide ID (e.g. 7)

# 2. Upload image via web UI slide card

# 3. Tile from local machine
go run ./cmd/tile-pipeline \
  -slide 7 \
  -magnification 40 \
  -microns-per-pixel 0.2505

# 4. Open viewer
open http://localhost:8080/slides/7/viewer
```

---

## Related documentation

- [Virtual Microscope — Technical Design](../VIRTUAL-MICROSCOPE-TECH.md)
- [Requirements — Virtual Microscope Viewer](../REQUIREMENTS.md#virtual-microscope-viewer)
- [Requirements Glossary — Spatial Calibration](../REQUIREMENTS-GLOSSARY.md#spatial-calibration)
