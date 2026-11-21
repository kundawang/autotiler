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

package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/krylphi/autotiler/internal/unpack"
)

const (
	inKey      = "in"
	paddingKey = "p"
	exportKey  = "e"
	outKey     = "o"
	tiledKey   = "j"
	forceKey   = "f"
)

const (
	export16  = "16"
	export28  = "28"
	export48  = "48"
	exportAll = "all"
)

func main() {
	args := parseArgs()
	inFiles, ok := args[inKey]
	if !ok {
		log.Print("Missing input file")
		os.Exit(1)
	}
	outFiles, outs := args[outKey]
	for i, inFile := range inFiles {
		inputFile := inFile
		outputFile := fmt.Sprintf("%d.local.png", i)
		if outs {
			if len(outFiles) > i {
				outputFile = outFiles[i]
			}
		}
		err := export(args, inputFile, outputFile)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
	}
}

func export(args map[string][]string, inputFile, outputFile string) error { // todo wrap errors
	cleanOut, err := safeOutputPath(outputFile)
	if err != nil {
		return err
	}

	imgFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("cannot open source image %q: %w", inputFile, err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("cannot decode source image %q: %w", inputFile, err)
	}

	padding := 0
	if paddings, ok := args[paddingKey]; ok {
		padding, err = strconv.Atoi(paddings[0])
		if err != nil {
			return err
		}
	}

	tiledOn := true
	if tiled, ok := args[tiledKey]; ok {
		tiledOn, err = parseBool(tiled[0])
		if err != nil {
			return fmt.Errorf("invalid -%s value: %w", tiledKey, err)
		}
	}

	force := false
	if forces, ok := args[forceKey]; ok {
		force, err = parseBool(forces[0])
		if err != nil {
			return fmt.Errorf("invalid -%s value: %w", forceKey, err)
		}
	}

	unpacker := unpack.NewUnpacker(img, 2, 3, padding)
	if err := unpacker.Init(2); err != nil {
		return err
	}

	exports, ok := args[exportKey]
	var exportTypes []string
	if !ok || len(exports) == 0 || exports[0] == exportAll {
		exportTypes = []string{export16, export28, export48}
	} else {
		exportTypes = exports
	}

	for e := range exportTypes {
		exportType := exportTypes[e]
		switch exportType {
		case export16:
			err := produceTileset(unpacker, unpacker.From6to16Terrain1, cleanOut, "16x1_terrain1", tiledOn, force)
			if err != nil {
				return err
			}
			err = produceTileset(unpacker, unpacker.From6to16Terrain2, cleanOut, "16x1_terrain2", tiledOn, force)
			if err != nil {
				return err
			}
		case export28:
			err := produceTileset(unpacker, unpacker.From6to28, cleanOut, "14x2", tiledOn, force)
			if err != nil {
				return err
			}
		case export48:
			err := produceTileset(unpacker, unpacker.From6to48Terrain1, cleanOut, "12x4_terrain1", tiledOn, force)
			if err != nil {
				return err
			}
			err = produceTileset(unpacker, unpacker.From6to48Terrain2, cleanOut, "12x4_terrain2", tiledOn, force)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseArgs() map[string][]string {
	if len(os.Args) < 2 {
		log.Print(
			"Usage: autotiler -in <file_in> [-o <file_out>] [-p <padding>] [-e <export_type(16,28,48,all)>]\n" +
				"                 [-j <tiled_json(1,0)>] [-f <force_overwrite(1,0)>]\n" +
				"       -e can be repeated\n" +
				"       -j defaults to 1: a Tiled JSON tileset is written next to every PNG\n" +
				"       -f defaults to 0: existing Tiled JSON files are not overwritten\n")
		os.Exit(1)
	}
	res := make(map[string][]string)
	allTilesets := false
	for i := 1; i < len(os.Args); i += 2 {
		key := strings.TrimPrefix(os.Args[i], "-")
		v, ok := res[key]
		value := os.Args[i+1]
		if key == exportKey && allTilesets {
			continue
		}
		if key == exportKey && strings.EqualFold(value, exportAll) {
			allTilesets = true
			v = []string{"all"}
			res[key] = v
			continue
		}
		if ok {
			v = append(v, value)
		} else {
			v = []string{value}
		}
		res[key] = v
	}
	return res
}

func produceTileset(unpacker *unpack.Unpacker, unpackFunction func() (*unpack.Sheet, error),
	outputPath, exportType string, tiledOn, force bool) error {
	sheet, err := unpackFunction()
	if err != nil {
		return err
	}
	cleanPath := filepath.Clean(outputPath)
	outputFile := fmt.Sprintf("%s_%s", exportType, filepath.Base(cleanPath))
	outputFile = filepath.Join(filepath.Dir(cleanPath), outputFile)
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)
	err = png.Encode(file, sheet.Canvas)
	if err != nil {
		return err
	}
	if !tiledOn {
		return nil
	}
	return produceTiledJSON(unpacker, sheet, outputFile, exportType, force)
}

// produceTiledJSON writes a Tiled JSON tileset next to the exported PNG.
// The JSON file shares the PNG's base name with a .json extension and
// references the PNG by its path relative to the JSON location.
func produceTiledJSON(unpacker *unpack.Unpacker, sheet *unpack.Sheet, pngPath, exportType string, force bool) error {
	jsonPath := strings.TrimSuffix(pngPath, filepath.Ext(pngPath)) + ".json"
	imageRef, err := filepath.Rel(filepath.Dir(jsonPath), pngPath)
	if err != nil {
		return fmt.Errorf("cannot resolve image path for %q: %w", jsonPath, err)
	}
	imageRef = filepath.ToSlash(imageRef)
	name := strings.TrimSuffix(filepath.Base(pngPath), filepath.Ext(pngPath))
	tileset := unpack.NewTiledTileset(name, imageRef, sheet,
		unpacker.TileWidth(), unpacker.TileHeight(), unpacker.Padding())
	return unpack.WriteTiledJSON(jsonPath, tileset, force)
}

// safeOutputPath cleans the given output path and rejects paths that would
// escape the current working directory.
func safeOutputPath(path string) (string, error) {
	clean := filepath.Clean(path)
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("output path %q escapes the working directory", path)
	}
	return clean, nil
}

// parseBool parses a boolean flag value: 1/0, true/false, on/off, yes/no.
func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "1", "true", "on", "yes":
		return true, nil
	case "0", "false", "off", "no":
		return false, nil
	}
	return false, fmt.Errorf("expected one of 1/0, true/false, on/off, yes/no, got %q", value)
}

//nolint:unused //debug function
func printArgs(args map[string][]string) {
	for key, params := range args {
		for _, param := range params {
			log.Printf("%s: %s\n", key, param)
		}
	}
}
