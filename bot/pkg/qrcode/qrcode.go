package qr

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
)

type Config struct {
	Content         string
	LogoPath        string
	Size            int
	LogoScale       float64
	Smoothing       float64 // Controls the overall smoothness of the QR code
	Background      color.Color
	Foreground      color.Color
	CornerRadius    float64 // Controls individual dot roundness
	RecoveryLevel   int
	QuietZone       int // Size of quiet zone around QR code
	LogoBackground  color.Color
	LogoBorderWidth float64 // Width of logo border
	LogoFade        float64 // Logo fade effect
}

// Generate creates a QR code with the given configuration and returns it as a byte slice
func (c *Config) Generate() ([]byte, error) {
	// Generate QR code with specified recovery level
	qr, err := qrcode.New(c.Content, qrcode.RecoveryLevel(c.RecoveryLevel))
	if err != nil {
		return nil, err
	}

	// Calculate total size including quiet zone (if enabled)
	totalSize := c.Size
	quietZoneOffset := 0
	if c.QuietZone > 0 {
		quietZoneOffset = c.QuietZone
		totalSize += 2 * c.QuietZone
	}

	// Create QR image with extra size for smoothing
	scaleFactor := 1 + c.Smoothing
	tempSize := int(float64(c.Size) * scaleFactor)
	qrImage := qr.Image(tempSize)

	// Create new context with total size including quiet zone
	dc := gg.NewContext(totalSize, totalSize)

	// Draw background
	dc.SetColor(c.Background)
	dc.Clear()

	// Create QR mask with fade effect if enabled
	qrMask := gg.NewContext(totalSize, totalSize)
	qrMask.SetColor(color.Black)

	// Calculate the fade area
	innerSize := float64(c.Size)
	if c.QuietZone > 0 {
		innerSize = float64(c.Size) - float64(c.QuietZone)*2
	}

	// Draw gradient from center to edges
	for y := 0; y < totalSize; y++ {
		for x := 0; x < totalSize; x++ {
			px := float64(x) - float64(totalSize)/2
			py := float64(y) - float64(totalSize)/2

			// Calculate distance from edge of QR code
			distanceFromEdge := math.Min(
				math.Min(
					math.Abs(math.Abs(px)-innerSize/2),
					math.Abs(math.Abs(py)-innerSize/2),
				),
				math.Min(
					innerSize/2-math.Abs(px),
					innerSize/2-math.Abs(py),
				),
			)

			if distanceFromEdge < 0 {
				// За пределами QR кода
				qrMask.SetRGBA(0, 0, 0, 0)
			} else {
				// Внутри QR кода
				qrMask.SetRGBA(0, 0, 0, 1)
			}
			qrMask.SetPixel(x, y)
		}
	}

	// Calculate logo dimensions and create mask if logo is provided
	logoSize := 0
	var logo image.Image
	var logoMask *gg.Context

	if c.LogoPath != "" {
		var err error
		logo, err = gg.LoadImage(c.LogoPath)
		if err != nil {
			return nil, err
		}
		logoSize = int(float64(c.Size) * c.LogoScale)

		// Create circular gradient mask for logo area
		logoMask = gg.NewContext(totalSize, totalSize)
		centerX := float64(totalSize) / 2
		centerY := float64(totalSize) / 2

		// Радиус области под лого
		logoRadius := float64(logoSize) / 2
		// Радиус области плавного перехода
		fadeRadius := logoRadius * (1 + c.LogoFade)

		// Создаем градиент от центра
		for y := 0; y < totalSize; y++ {
			for x := 0; x < totalSize; x++ {
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance < logoRadius {
					// Полностью прозрачная область под лого
					logoMask.SetRGBA(0, 0, 0, 0)
				} else if distance < fadeRadius {
					// Плавный переход
					progress := (distance - logoRadius) / (fadeRadius - logoRadius)
					alpha := math.Min(1, progress)
					logoMask.SetRGBA(0, 0, 0, alpha)
				} else {
					// Полностью непрозрачная область
					logoMask.SetRGBA(0, 0, 0, 1)
				}
				logoMask.SetPixel(x, y)
			}
		}
	}

	// Calculate QR code matrix size
	qrMatrix := qr.Bitmap()
	matrixSize := len(qrMatrix)

	// Calculate scaling factors
	scaleX := float64(c.Size) / float64(tempSize)
	scaleY := float64(c.Size) / float64(tempSize)

	// Draw smoothed QR code
	bounds := qrImage.Bounds()
	dotSize := float64(c.Size) * c.CornerRadius

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Convert pixel coordinates to matrix coordinates
			matrixX := int(float64(x) / float64(tempSize) * float64(matrixSize))
			matrixY := int(float64(y) / float64(tempSize) * float64(matrixSize))

			// Skip if coordinates are invalid
			if matrixX >= matrixSize || matrixY >= matrixSize {
				continue
			}

			r, _, _, _ := qrImage.At(x, y).RGBA()
			if r == 0 { // black pixel
				px := float64(x)*scaleX + float64(quietZoneOffset)
				py := float64(y)*scaleY + float64(quietZoneOffset)

				// Get alpha from QR mask
				_, _, _, qrAlpha := qrMask.Image().At(int(px), int(py)).RGBA()
				qrFade := float64(qrAlpha) / 65535.0

				// Если есть лого, проверяем маску в этой точке
				if logo != nil {
					maskX := int(px)
					maskY := int(py)
					if maskX >= 0 && maskX < totalSize && maskY >= 0 && maskY < totalSize {
						if logoMask == nil {
							continue
						}
						_, _, _, a := logoMask.Image().At(maskX, maskY).RGBA()
						alpha := float64(a) / 65535.0 // Нормализуем значение альфа-канала

						if alpha < 1 {
							// Если точка находится в области маски, пропускаем или делаем частично прозрачной
							if alpha == 0 {
								continue
							}
							alpha = math.Min(alpha, qrFade)
							dc.SetRGBA(
								float64(c.Foreground.(color.RGBA).R)/255.0,
								float64(c.Foreground.(color.RGBA).G)/255.0,
								float64(c.Foreground.(color.RGBA).B)/255.0,
								alpha,
							)
						} else {
							dc.SetRGBA(
								float64(c.Foreground.(color.RGBA).R)/255.0,
								float64(c.Foreground.(color.RGBA).G)/255.0,
								float64(c.Foreground.(color.RGBA).B)/255.0,
								qrFade,
							)
						}
					}
				} else {
					dc.SetRGBA(
						float64(c.Foreground.(color.RGBA).R)/255.0,
						float64(c.Foreground.(color.RGBA).G)/255.0,
						float64(c.Foreground.(color.RGBA).B)/255.0,
						qrFade,
					)
				}

				dc.DrawCircle(px, py, dotSize)
				dc.Fill()
			}
		}
	}

	// If logo path is provided, embed the logo
	if logo != nil {
		// Resize logo
		resizedLogo := resize.Resize(uint(logoSize), uint(logoSize), logo, resize.Lanczos3)

		// Create a new context for the circular logo
		logoCtx := gg.NewContext(logoSize, logoSize)

		// Draw logo background circle with border
		if c.LogoBorderWidth > 0 {
			// Draw outer circle (border)
			logoCtx.SetColor(c.LogoBackground)
			logoCtx.DrawCircle(float64(logoSize)/2, float64(logoSize)/2, float64(logoSize)/2)
			logoCtx.Fill()

			// Draw inner circle (actual logo background)
			logoCtx.SetColor(c.LogoBackground)
			logoCtx.DrawCircle(float64(logoSize)/2, float64(logoSize)/2, float64(logoSize)/2-c.LogoBorderWidth)
			logoCtx.Fill()
		} else {
			logoCtx.SetColor(c.LogoBackground)
			logoCtx.DrawCircle(float64(logoSize)/2, float64(logoSize)/2, float64(logoSize)/2)
			logoCtx.Fill()
		}

		// Draw logo
		logoCtx.DrawImage(resizedLogo, 0, 0)

		// Create circular mask
		maskCtx := gg.NewContext(logoSize, logoSize)
		maskCtx.SetColor(c.Foreground)
		maskCtx.DrawCircle(float64(logoSize)/2, float64(logoSize)/2, float64(logoSize)/2)
		maskCtx.Fill()

		// Apply mask to logo
		for x := 0; x < logoSize; x++ {
			for y := 0; y < logoSize; y++ {
				_, _, _, ma := maskCtx.Image().At(x, y).RGBA()
				if ma == 0 {
					logoCtx.SetColor(color.Transparent)
					logoCtx.SetPixel(x, y)
				}
			}
		}

		// Draw the circular logo at center
		x := (totalSize - logoSize) / 2
		y := (totalSize - logoSize) / 2
		dc.DrawImage(logoCtx.Image(), x, y)
	}

	// Convert to byte slice
	var buf bytes.Buffer
	err = png.Encode(&buf, dc.Image())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
