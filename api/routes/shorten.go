package routes

import (
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}
 
type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest int           `json:"rate_limit_reset"`
}

// removeDomainError → تمنع إن المستخدم يحط نفس الدومين بتاعك
func removeDomainError(url string, domain string) bool {
	return !strings.Contains(url, domain)
}

// enforceHTTPS → يحول أي URL لـ HTTPS
func enforceHTTPS(url string) string {
	if strings.HasPrefix(url, "http://") {
		return strings.Replace(url, "http://", "https://", 1)
	}
	if !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

// Shorten handles the URL shortening requests
func Shorten(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// check if the input url is valid
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// check for domain error (replace localhost:3000 بالدومين بتاعك)
	if !removeDomainError(body.URL, "localhost:3000") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You can't shorten this domain",
		})
	}

	// enforce https , SSL
	body.URL = enforceHTTPS(body.URL)

	return c.Status(fiber.StatusOK).JSON(response{
		URL:            body.URL,
		CustomShort:    body.CustomShort,
		Expiry:         body.Expiry,
		XRateRemaining: 10,
		XRateLimitRest: 10,
	})
}
