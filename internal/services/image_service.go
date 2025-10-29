package services

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"countryCurrency/internal/database"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type ImageService struct {
	repo      *database.Repository
	imagePath string
}

func NewImageService(repo *database.Repository, imagePath string) *ImageService {
	return &ImageService{
		repo:      repo,
		imagePath: imagePath,
	}
}

func (s *ImageService) GenerateSummaryImage() error {
	totalCountries, err := s.repo.GetTotalCountries()
	if err != nil {
		return fmt.Errorf("failed to get total countries: %w", err)
	}

	topCountries, err := s.repo.GetTopCountriesByGDP(5)
	if err != nil {
		return fmt.Errorf("failed to get top countries: %w", err)
	}

	lastRefresh, err := s.repo.GetLastRefreshedAt()
	if err != nil {
		return fmt.Errorf("failed to get last refresh time: %w", err)
	}

	width, height := 600, 400
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	s.fillBackground(img, color.White)

	y := 30
	y = s.drawText(img, fmt.Sprintf("Total countries: %d", totalCountries), 20, y, color.Black)
	y += 20
	y = s.drawText(img, "Top 5 countries by GDP:", 20, y, color.Black)
	y += 10

	for i, country := range topCountries {
		gdp := "N/A"
		if country.EstimatedGDP != nil {
			gdp = fmt.Sprintf("$%.2f", *country.EstimatedGDP)
		}
		text := fmt.Sprintf("%d. %s - %s", i+1, country.Name, gdp)
		y = s.drawText(img, text, 40, y, color.RGBA{50, 50, 50, 255})
		y += 5
	}

	y += 20
	s.drawText(img, fmt.Sprintf("last refreshed: %s", lastRefresh.Format("2006-01-02 15:04:05")), 20, y, color.RGBA{100, 100, 100, 255})

	if err := os.MkdirAll(filepath.Dir(s.imagePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	file, err := os.Create(s.imagePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

func (s *ImageService) fillBackground(img *image.RGBA, col color.Color) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, col)
		}
	}
}

func (s *ImageService) drawText(img *image.RGBA, text string, x, y int, col color.Color) int {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)

	return y + 20
}

func (s *ImageService) GetImagePath() string {
	return s.imagePath
}
