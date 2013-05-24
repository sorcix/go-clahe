// CLAHE implemented in Go
//
// Original CLAHE implementation by Karel Zuiderveld, karel@cv.ruu.nl
// in "Graphics Gems IV", Academic Press, 1994.
//
// Written in Go by Vic Demuzere, vic@demuzere.be
//
// License: MIT - http://opensource.org/licenses/MIT
//

package clahe

import (

	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"sync"
)

// Picture represents an image.
type Picture struct {

	File *image.Gray

	Width  int
	Height int

	BinCount uint8

	ClipLimit int

	ColorMax uint8
	ColorMin uint8

	BlockCountX int
	BlockCountY int
	BlockWidth  int
	BlockHeight int

	Pixels [][]uint8

	Blocks [][]*Block

	LUT []uint8

	WaitGroup *sync.WaitGroup
}

// Read loads an image from disk.
func (picture *Picture) Read(path string) error {

	// Open image file
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Shedule close
	defer file.Close()

	// Decode image file
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	var oldColor color.Color
	var newColor color.Gray

	// Image properties
	bounds := img.Bounds()
	picture.Width, picture.Height = bounds.Max.X, bounds.Max.Y
	picture.ColorMin = 255
	picture.ColorMax = 0

	// Convert to grayscale
	picture.File = image.NewGray(bounds)
	for x := 0; x < picture.Width; x++ {
		for y := 0; y < picture.Height; y++ {
			oldColor = img.At(x, y)
			newColor = color.GrayModel.Convert(oldColor).(color.Gray)

			switch {
			case newColor.Y < picture.ColorMin:
				picture.ColorMin = newColor.Y
			case newColor.Y > picture.ColorMax:
				picture.ColorMax = newColor.Y
			}

			picture.File.Set(x, y, newColor)
		}
	}

	// Pointer magic!
	offset := 0
	picture.Pixels = make([][]uint8, picture.Width)
	for y := 0; y < picture.Height; y++ {
		offset = y * picture.Width
		picture.Pixels[y] = picture.File.Pix[offset:offset+picture.Width]
	}

	return nil

}

func (picture *Picture) Write(path string) error {

	// Open image file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	
	// Shedule close
    defer file.Close()

	// Encode!
    err = png.Encode(file, picture.File)
    
    return err
    
}

// GenerateLUT calculates the color lookup table.
func (picture *Picture) GenerateLUT() {

	picture.LUT = make([]uint8, 256)
	binSize := 1 + ((picture.ColorMax - picture.ColorMin) / picture.BinCount)
	for i := picture.ColorMin; i < picture.ColorMax; i++ {
		picture.LUT[i] = (i - picture.ColorMin) / binSize
	}

}

// CLAHE improves contrast on the picture.
func (picture *Picture) CLAHE(blockCountX, blockCountY int, clipLimit float32) {

	if picture.BinCount < 128 {
		picture.BinCount = 128
	}

	picture.BlockCountX = blockCountX
	picture.BlockCountY = blockCountY

	// Store blocksizes as we'll need them a lot!
	picture.BlockWidth = picture.Width / picture.BlockCountX
	picture.BlockHeight = picture.Height / picture.BlockCountY

	// Calculate absolute cliplimit
	picture.ClipLimit = int(clipLimit * float32((picture.BlockWidth*picture.BlockHeight)/int(picture.BinCount)))

	// Generate lookup table
	picture.GenerateLUT()
	
	offset := 0

	// Prepare blocks
	picture.Blocks = make([][]*Block, picture.BlockCountX)
	for y := 0; y < picture.BlockCountY; y++ {

		picture.Blocks[y] = make([]*Block, picture.BlockCountX)

		for x := 0; x < picture.BlockCountX; x++ {
			picture.Blocks[y][x] = new(Block)

			// Pointer magic!
			picture.Blocks[y][x].Pixels = make([][]uint8, picture.BlockWidth)
			for i := 0; i < picture.BlockHeight; i++ {
				offset = x*picture.BlockWidth
				picture.Blocks[y][x].Pixels[i] = picture.Pixels[picture.BlockHeight*y+i][offset:offset+picture.BlockWidth]
			}
			
			picture.Blocks[y][x].Picture = picture
		}
	}
	
	picture.WaitGroup = new(sync.WaitGroup)
	
	// Prepare interpolation
	picture.PrepareInterpolation()
	
	// Generate histograms!
	for x := 0; x < picture.BlockCountX; x++ {
		for y := 0; y < picture.BlockCountY; y++ {
			go picture.Blocks[y][x].CalculateHistogram(x,y)
		}
	}
	
	// Wait for interpolation to finish.
	picture.WaitGroup.Wait();

}

func (picture *Picture) PrepareInterpolation() {

	var top, bottom, left, right, subWidth, subHeight, offsetX, offsetY int

	for blockY := 0; blockY <= picture.BlockCountY; blockY++ {
		offsetX = 0

		switch blockY {
		case 0:
			// TOP ROW
			subHeight = picture.BlockHeight / 2
			top = 0
			bottom = 0
		case picture.BlockCountY:
			// BOTTOM ROW
			subHeight = picture.BlockHeight / 2
			top = picture.BlockCountY - 1
			bottom = top
		default:
			subHeight = picture.BlockHeight
			top = blockY - 1
			bottom = blockY
		}

		for blockX := 0; blockX <= picture.BlockCountX; blockX++ {
			switch blockX {
			case 0:
				// LEFT COLUMN
				subWidth = picture.BlockWidth / 2
				left = 0
				right = 0
			case picture.BlockCountX:
				// RIGHT COLUMN
				subWidth = picture.BlockWidth / 2
				left = picture.BlockCountX - 1
				right = left
			default:
				subWidth = picture.BlockWidth
				left = blockX - 1
				right = blockX
			}

			subBlock := new(SubBlock)

			// Properties
			subBlock.Width = subWidth
			subBlock.Height = subHeight
			subBlock.OffsetX = offsetX
			subBlock.OffsetY = offsetY
			
			subBlock.Picture = picture

			// This subblock depends on 4 blocks!
			subBlock.TopLeft = picture.Blocks[top][left]
			subBlock.TopRight = picture.Blocks[top][right]
			subBlock.BottomLeft = picture.Blocks[bottom][left]
			subBlock.BottomRight = picture.Blocks[bottom][right]

			// We expect 4 blocks
			subBlock.WaitGroup = new(sync.WaitGroup)
			subBlock.WaitGroup.Add(4)

			// Ask for notification from the 4 blocks we need to continue.
			subBlock.TopLeft.PleaseNotify(subBlock)
			subBlock.TopRight.PleaseNotify(subBlock)
			subBlock.BottomLeft.PleaseNotify(subBlock)
			subBlock.BottomRight.PleaseNotify(subBlock)

			// Shedule interpolation for this block.
			go subBlock.Interpolate()

			// Picture has one more subblock to wait for.
			picture.WaitGroup.Add(1)

			// Offset in image klaarzetten voor volgende loop!
			offsetX += subWidth
		}
		offsetY += subHeight
	}

	//Task.WaitAll(tasks.ToArray());
}
