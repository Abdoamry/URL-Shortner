package routes

import (
	"strings"
	"time"
	"os"
	"strconv"

	"github.com/Abdoamry/URL-Shortner/database"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

// request → شكل البيانات اللي المستخدم هيبعتها
type request struct {
	URL         string        `json:"url"`   // الرابط الأصلي اللي عايزين نعمله Shorten
	CustomShort string        `json:"short"` // اختيار اسم مخصص للرابط المختصر
	Expiry      time.Duration `json:"expiry"`// وقت انتهاء صلاحية الرابط
}

// response → شكل البيانات اللي هنرجعها للمستخدم
type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`       // عدد المحاولات المتبقية
	XRateLimitRest int           `json:"rate_limit_reset"` // وقت إعادة التعيين
}

// removeDomainError → تمنع إن المستخدم يحط نفس الدومين بتاع الخدمة
func removeDomainError(url string, domain string) bool {
	return !strings.Contains(url, domain)
}

// enforceHTTPS → يحول أي URL لـ HTTPS (لو مفيش بروتوكول بيضيفه)
func enforceHTTPS(url string) string {
	if strings.HasPrefix(url, "http://") {
		return strings.Replace(url, "http://", "https://", 1)
	}
	if !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

// Shorten → الـ handler الأساسي لعمل Shorten للرابط
func Shorten(c *fiber.Ctx) error {
	// نحط البودي اللي جاي من الـ request جوه struct request
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// --- Rate Limiting ---
	r2 := database.CreateClient(1) // نستخدم DB رقم 1 في Redis عشان rate limiting
	defer r2.Close()

	// نشوف الـ IP بتاع المستخدم موجود ولا لأ
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// أول مرة → نحطله قيمة (عدد المحاولات المسموحة)
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QOUTE"), 30*60*time.Second).Err()
	} else {
		// لو موجود → نتاكد إن لسه ليه محاولات
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"reset": limit.Seconds(), // هيرجع الوقت بالثواني لحد ما يتصفر
			})
		}
	}

	// --- Validate Input URL ---
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// --- Check Domain Error ---
	if !removeDomainError(body.URL, "localhost:3000") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You can't shorten this domain",
		})
	}

	// --- Enforce HTTPS ---
	body.URL = enforceHTTPS(body.URL)

	// قلل عدد المحاولات المتبقية
	r2.Decr(database.Ctx, c.IP())

	// --- Response ---
	return c.Status(fiber.StatusOK).JSON(response{
		URL:            body.URL,
		CustomShort:    body.CustomShort,
		Expiry:         body.Expiry,
		XRateRemaining: 10, // هنا لسه محتاج نجيب القيمة الصح من Redis
		XRateLimitRest: 1800, // بالثواني (نص ساعة) - ممكن نجيبها ديناميكياً من TTL
	})
}
