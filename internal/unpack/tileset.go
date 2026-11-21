/*
 * MIT License
 *
 * Copyright (c) 2024 The autotiler authors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package unpack

import (
	"image"
	"image/draw"
)

// TileMeta describes the provenance of a single output tile: which sub tiles
// of the source packed tileset it was composed from and how the composed tile
// was rotated when placed on the canvas.
type TileMeta struct {
	// Quads are the source sub tile coordinates in top-left, top-right,
	// bottom-left, bottom-right order, before any rotation was applied.
	Quads [4][2]int
	// Rotation is the cumulative counter-clockwise rotation in degrees
	// (0, 90, 180 or 270) applied to the composed tile when it was placed.
	Rotation int
}

// Sheet is a generated tileset canvas together with the layout information
// and per-tile provenance required to describe it to external tools
// (e.g. a Tiled tileset).
type Sheet struct {
	Canvas  *image.NRGBA
	Columns int
	Rows    int
	// Tiles holds the provenance of every tile on the canvas in row-major
	// order, so Tiles[i] describes the tile with ID i.
	Tiles []TileMeta
}

type tileSet struct {
	canvas                *image.NRGBA
	tileWidth, tileHeight int
	columns               int
	placements            map[int]TileMeta
}

func newTileSet(canvas *image.NRGBA, tileWidth, tileHeight, columns int) *tileSet {
	return &tileSet{
		canvas:     canvas,
		tileWidth:  tileWidth,
		tileHeight: tileHeight,
		columns:    columns,
		placements: make(map[int]TileMeta),
	}
}

// setTile draws the given tile onto the canvas at the specified coordinates (x, y).
// The tile is drawn at its original orientation.
//
// Parameters:
// - x: The x-coordinate on the canvas where the tile will be placed.
// - y: The y-coordinate on the canvas where the tile will be placed.
// - tile: The image.NRGBA tile to be drawn onto the canvas.
// - meta: The provenance of the tile being placed, recorded for traceability.
func (t *tileSet) setTile(x, y int, tile *image.NRGBA, meta TileMeta) {
	// Calculate the minimum point of the area on the canvas where the tile will be drawn.
	canvasMin := image.Point{
		X: x * t.tileWidth,
		Y: y * t.tileHeight,
	}

	// Define the area on the canvas where the tile will be drawn.
	canvasArea := image.Rectangle{
		Min: canvasMin,
		Max: image.Point{
			X: canvasMin.X + t.tileWidth,
			Y: canvasMin.Y + t.tileHeight,
		},
	}

	// Draw the tile onto the canvas at the specified area.
	draw.Draw(
		t.canvas,
		canvasArea,
		tile,
		image.Point{
			X: 0,
			Y: 0,
		},
		draw.Src,
	)

	// Record the provenance of the placed tile using its row-major tile ID.
	t.placements[y*t.columns+x] = meta
}

// setTileWithRotationLeft draws the given tile onto the canvas at the specified coordinates (x, y)
// and rotates the tile 90 degrees to the left before drawing it.
// The tile is drawn at its rotated orientation.
//
// Parameters:
//   - x: The x-coordinate on the canvas where the tile will be placed.
//   - y: The y-coordinate on the canvas where the tile will be placed.
//   - tile: The image.NRGBA tile to be drawn onto the canvas.
//   - meta: The provenance of the tile being placed, recorded for traceability.
//     The caller is expected to have accounted for the additional rotation in meta.Rotation.
//
// Returns:
// - A pointer to the rotated image.NRGBA tile.
func (t *tileSet) setTileWithRotationLeft(x, y int, tile *image.NRGBA, meta TileMeta) *image.NRGBA {
	// Rotate the tile 90 degrees to the left.
	rotatedTile := rotateLeft90(tile)

	// Draw the rotated tile onto the canvas at the specified coordinates.
	t.setTile(x, y, rotatedTile, meta)

	// Return the rotated tile.
	return rotatedTile
}

func (t *tileSet) getCanvas() *image.NRGBA {
	return t.canvas
}

// getPlacements returns the recorded per-tile provenance as a slice indexed
// by row-major tile ID. Positions that were never placed keep a zero TileMeta.
func (t *tileSet) getPlacements(count int) []TileMeta {
	tiles := make([]TileMeta, count)
	for id, meta := range t.placements {
		if id >= 0 && id < count {
			tiles[id] = meta
		}
	}
	return tiles
}
