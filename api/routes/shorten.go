package routes

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Abdoamry/URL-Shortner/database"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func removeDomainError(url string, domain string) bool {
	return !strings.Contains(url, domain)
}

func enforceHTTPS(url string) string {
	if strings.HasPrefix(url, "http://") {
		return strings.Replace(url, "http://", "https://", 1)
	}
	if !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func Shorten(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Rate Limiting
	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err != nil && err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error checking rate limit",
		})
	}
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QOUTE"), 30*60*time.Second).Err()
	} else {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid rate limit value",
			})
		}
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"reset": limit.Seconds(),
			})
		}
	}

	// Validate URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// Check Domain
	if !removeDomainError(body.URL, "localhost:3000") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You can't shorten this domain",
		})
	}

	// Enforce HTTPS
	body.URL = enforceHTTPS(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
		val, _ := database.CreateClient(0).Get(database.Ctx, id).Result()
		if val != "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Custom short already exists",
			})
		}
	}

	// Set Expiry
	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}

	r := database.CreateClient(0)
	defer r.Close()
	err = r.Set(database.Ctx, id, body.URL, body.Expiry).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Decrease rate limit
	r2.Decr(database.Ctx, c.IP())

	// Dynamic response
	remaining, _ := strconv.Atoi(r2.Get(database.Ctx, c.IP()).Val())  
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	return c.Status(fiber.StatusOK).JSON(response{
		URL:            body.URL,
		CustomShort:    id,
		Expiry:         body.Expiry,
		XRateRemaining: remaining,
		XRateLimitRest: int(ttl.Seconds()),
	})

	resp := response{
		URL:            body.URL,
		CustomShort:    id,
		Expiry:         body.Expiry,
		XRateRemaining: remaining,
		XRateLimitRest: int(ttl.Seconds()),
	}

	val ,_ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining,_= strconv.Atoi(val)

	ttl, _ = r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitRest = int(ttl.Seconds())
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)
}