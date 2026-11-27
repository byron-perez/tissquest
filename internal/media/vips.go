package media

import (
	"fmt"
	"os/exec"
	"strconv"
)

func Thumbnail(src, dst string, width, height, quality int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid thumbnail dimensions %dx%d", width, height)
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}

	output := fmt.Sprintf("%s[Q=%d]", dst, quality)
	cmd := exec.Command(
		"vips", "thumbnail", src, output, strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--size", "down",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vips thumbnail failed: %w: %s", err, string(out))
	}
	return nil
}
