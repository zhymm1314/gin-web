package fx

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/fx"

	"gin-web/app/controllers"
	"gin-web/config"
)

// 颜色定义
var (
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	green   = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	white   = color.New(color.FgWhite, color.Bold).SprintFunc()
	gray    = color.New(color.FgHiBlack).SprintFunc()
)

const banner = `
   ______    _           _       __         __
  / ____/   (_)   ____  | |     / /  ___   / /_
 / / __    / /   / __ \ | | /| / /  / _ \ / __ \
/ /_/ /   / /   / / / / | |/ |/ /  /  __// /_/ /
\____/   /_/   /_/ /_/  |__/|__/   \___/ /_.___/
`

// BannerModule 启动信息模块
var BannerModule = fx.Module("banner",
	fx.Invoke(PrintBanner),
)

// ModuleStatus 模块状态
type ModuleStatus struct {
	Name    string
	Enabled bool
}

// PrintBanner 打印启动信息
func PrintBanner(cfg *config.Configuration, params ControllerParams) {
	// 打印 Banner
	fmt.Println(cyan(banner))

	// 分隔线
	printDivider()

	// 基本信息
	fmt.Println(white(" Application Info"))
	printDivider()
	printKeyValue("Name", cfg.App.AppName)
	printKeyValue("Environment", getEnvBadge(cfg.App.Env))
	printKeyValue("Port", cfg.App.Port)
	printKeyValue("Go Version", runtime.Version())
	printKeyValue("Start Time", time.Now().Format("2006-01-02 15:04:05"))

	// 模块状态
	printDivider()
	fmt.Println(white(" Modules"))
	printDivider()
	modules := []ModuleStatus{
		{"Database (MySQL)", true},
		{"Redis", true},
		{"RabbitMQ", cfg.RabbitMQ.Enable},
		{"Cron Jobs", cfg.Cron.Enable},
		{"WebSocket", cfg.WebSocket.Enable},
	}
	for _, m := range modules {
		printModuleStatus(m.Name, m.Enabled)
	}

	// 注册的控制器
	printDivider()
	fmt.Println(white(" Registered Controllers"))
	printDivider()
	for _, ctrl := range params.Controllers {
		printController(ctrl)
	}

	// URL 信息
	printDivider()
	fmt.Println(white(" URLs"))
	printDivider()
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.App.Port)
	printKeyValue("API", baseURL+"/api")
	if cfg.App.Env != "production" {
		printKeyValue("Swagger", baseURL+"/swagger/index.html")
	}

	// 完成信息
	printDivider()
	fmt.Printf(" %s Server started successfully!\n", green("✓"))
	printDivider()
	fmt.Println()
}

// printDivider 打印分隔线
func printDivider() {
	fmt.Println(gray(" " + strings.Repeat("-", 50)))
}

// printKeyValue 打印键值对
func printKeyValue(key, value string) {
	fmt.Printf(" %s %s\n", blue(fmt.Sprintf("%-15s", key+":")), value)
}

// printModuleStatus 打印模块状态
func printModuleStatus(name string, enabled bool) {
	status := green("✓ Enabled")
	if !enabled {
		status = gray("○ Disabled")
	}
	fmt.Printf(" %-25s %s\n", name, status)
}

// printController 打印控制器信息
func printController(ctrl controllers.Controller) {
	prefix := ctrl.Prefix()
	routes := ctrl.Routes()
	fmt.Printf(" %s %s\n", magenta("►"), yellow("/api"+prefix))
	for _, route := range routes {
		method := formatMethod(route.Method)
		fmt.Printf("   %s %s\n", method, gray(route.Path))
	}
}

// formatMethod 格式化 HTTP 方法
func formatMethod(method string) string {
	switch method {
	case "GET":
		return color.New(color.FgGreen).Sprintf("%-7s", method)
	case "POST":
		return color.New(color.FgBlue).Sprintf("%-7s", method)
	case "PUT":
		return color.New(color.FgYellow).Sprintf("%-7s", method)
	case "DELETE":
		return color.New(color.FgRed).Sprintf("%-7s", method)
	case "PATCH":
		return color.New(color.FgMagenta).Sprintf("%-7s", method)
	default:
		return fmt.Sprintf("%-7s", method)
	}
}

// getEnvBadge 获取环境标签
func getEnvBadge(env string) string {
	switch env {
	case "production", "prod":
		return color.New(color.FgRed, color.Bold).Sprint("[PRODUCTION]")
	case "dev", "development":
		return color.New(color.FgGreen).Sprint("[DEV]")
	case "pre", "staging":
		return color.New(color.FgYellow).Sprint("[STAGING]")
	default:
		return color.New(color.FgCyan).Sprintf("[%s]", strings.ToUpper(env))
	}
}
