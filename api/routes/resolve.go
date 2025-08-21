package routes

import(
	"github.com/Abdoamry/URL-Shortner/database" // استدعاء ملف قاعدة البيانات (علشان نتعامل مع Redis)
	"github.com/go-redis/redis/v8"              // مكتبة Redis للتعامل مع قواعد البيانات
	"github.com/gofiber/fiber/v2"               // مكتبة Fiber (framework) علشان نبني REST API
)

func Resolve(c *fiber.Ctx) error {
	// هنا بناخد ال parameter من ال URL 
	// يعني لو المستخدم دخل /abcd فهنا url = "abcd"
	url := c.Params("url")

	// فتح connection مع Redis database رقم 0 (اللي بيخزن الـ short links)
	r := database.CreateClient(0)
	// لازم نقفل الـ connection بعد ما نخلص منه
	defer r.Close()

	// محاولة الحصول على الـ original URL من Redis عن طريق الـ key (اللي هو url)
	value ,err := r.Get(database.Ctx , url).Result()

	// لو الـ key مش موجود في Redis → نرجع رسالة Not Found
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "URL not found",
		})
	// لو حصل خطأ تاني (مش الـ key فاضي) → نرجع Internal Server Error
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
	
	// فتح connection مع Redis database رقم 1 (اللي بيخزن الـ analytics مثلاً زي عدد الزيارات)
	rInr := database.CreateClient(1)
	defer rInr.Close()
	
	// كل مرة المستخدم يزور الرابط القصير → نزود العداد بواحد
	_ = rInr.Incr(database.Ctx , "counter")

	// نعمل redirect للـ original URL
	// يعني لو الرابط القصير abcd → يفتح الرابط الأصلي مثلاً google.com
	return c.Redirect(value , fiber.StatusFound)
}
