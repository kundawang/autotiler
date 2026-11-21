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
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ErrOutputExists is returned when a Tiled tileset export would overwrite an
// existing file and overwriting was not explicitly allowed.
var ErrOutputExists = errors.New("output file already exists")

const (
	tiledVersion     = "1.10.2"
	tiledJSONVersion = "1.10"
)

// TiledProperty is a single custom property of a Tiled tile.
type TiledProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// TiledTile is a single tile entry of a Tiled tileset. The ID matches the
// row-major index of the tile on the exported sheet.
type TiledTile struct {
	ID         int             `json:"id"`
	Properties []TiledProperty `json:"properties,omitempty"`
}

// TiledTileset is a Tiled JSON tileset (https://doc.mapeditor.org/en/stable/reference/json-map-format/#tileset)
// describing one exported sheet.
type TiledTileset struct {
	Columns      int         `json:"columns"`
	Image        string      `json:"image"`
	ImageHeight  int         `json:"imageheight"`
	ImageWidth   int         `json:"imagewidth"`
	Margin       int         `json:"margin"`
	Name         string      `json:"name"`
	Spacing      int         `json:"spacing"`
	TileCount    int         `json:"tilecount"`
	TiledVersion string      `json:"tiledversion"`
	TileHeight   int         `json:"tileheight"`
	TileWidth    int         `json:"tilewidth"`
	Type         string      `json:"type"`
	Version      string      `json:"version"`
	Tiles        []TiledTile `json:"tiles,omitempty"`
}

// NewTiledTileset builds a Tiled JSON tileset description for the given sheet.
//
// Parameters:
//   - name: the tileset name recorded in the JSON.
//   - imagePath: the path to the sheet image as it should be referenced from the
//     JSON file (usually relative to the JSON location).
//   - sheet: the exported sheet the tileset describes.
//   - tileWidth, tileHeight: the unpadded size of a single tile in pixels.
//   - padding: the padding in pixels added around every tile on the sheet. It maps
//     to Tiled's margin (padding) and spacing (2*padding) fields.
//
// Returns:
// - A TiledTileset ready to be marshalled to JSON.
func NewTiledTileset(name, imagePath string, sheet *Sheet, tileWidth, tileHeight, padding int) *TiledTileset {
	tiles := make([]TiledTile, len(sheet.Tiles))
	for id, meta := range sheet.Tiles {
		tiles[id] = TiledTile{
			ID: id,
			Properties: []TiledProperty{
				{
					Name:  "quads",
					Type:  "string",
					Value: quadsString(meta.Quads),
				},
				{
					Name:  "rotation",
					Type:  "int",
					Value: meta.Rotation,
				},
			},
		}
	}
	return &TiledTileset{
		Columns:      sheet.Columns,
		Image:        imagePath,
		ImageHeight:  sheet.Canvas.Bounds().Dy(),
		ImageWidth:   sheet.Canvas.Bounds().Dx(),
		Margin:       padding,
		Name:         name,
		Spacing:      padding * 2,
		TileCount:    sheet.Columns * sheet.Rows,
		TiledVersion: tiledVersion,
		TileHeight:   tileHeight,
		TileWidth:    tileWidth,
		Type:         "tileset",
		Version:      tiledJSONVersion,
		Tiles:        tiles,
	}
}

// quadsString renders source sub tile coordinates as "x,y;x,y;x,y;x,y" in
// top-left, top-right, bottom-left, bottom-right order.
func quadsString(quads [4][2]int) string {
	return fmt.Sprintf("%d,%d;%d,%d;%d,%d;%d,%d",
		quads[0][0], quads[0][1],
		quads[1][0], quads[1][1],
		quads[2][0], quads[2][1],
		quads[3][0], quads[3][1],
	)
}

// WriteTiledJSON marshals the tileset and writes it to the given path.
// Unless force is true, it refuses to overwrite an existing file and returns
// an error wrapping ErrOutputExists.
func WriteTiledJSON(path string, ts *TiledTileset, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s (use -f 1 to overwrite)", ErrOutputExists, path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("cannot inspect output file %q: %w", path, err)
		}
	}
	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode Tiled tileset %q: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("cannot write Tiled tileset %q: %w", path, err)
	}
	return nil
}
