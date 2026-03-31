package routes

import (
	"gitee.com/dinglide/spot-vm/internal/config"
	"gitee.com/dinglide/spot-vm/internal/handlers"
	"gitee.com/dinglide/spot-vm/internal/middleware"
	"gitee.com/dinglide/spot-vm/internal/tcc"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(cfg *config.Config, tccClient *tcc.TCC) *gin.Engine {
	// 设置 Gin 模式
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 使用中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.AuditMiddleware())

	// 创建处理器
	authHandler := handlers.NewSimpleAuthHandler(cfg)
	spotVMHandler := handlers.NewSpotVMHandler(cfg, tccClient)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 健康检查（无需认证）
		v1.GET("/health", authHandler.Health)

		// API key验证（无需认证）
		v1.POST("/auth/validate", authHandler.ValidateAPIKey)

		// 需要认证的路由
		protected := v1.Group("/")
		protected.Use(middleware.APIKeyAuthMiddleware(cfg))
		{
			// 这里可以添加需要API key认证的路由
			protected.GET("/protected", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "This is a protected endpoint"})
			})

			// Spot VM 相关路由
			spotVM := protected.Group("/spot-vm")
			{
				// 获取最便宜的实例列表
				spotVM.GET("/cheapest", spotVMHandler.GetCheapestInstances)

				// 创建最便宜的Spot VM
				spotVM.POST("/create-cheapest", spotVMHandler.CreateCheapestSpotVM)

				// 获取可用区列表
				spotVM.GET("/zones", spotVMHandler.GetAvailableZones)

				// 获取指定可用区的实例类型
				spotVM.GET("/instance-types", spotVMHandler.GetInstanceTypes)

				// 删除实例
				spotVM.DELETE("/instances/:id", spotVMHandler.DeleteInstance)

				// 当前实例状态管理
				spotVM.GET("/current/status", spotVMHandler.GetCurrentInstanceStatus)
				spotVM.GET("/current/region", spotVMHandler.GetCurrentRegion)
				spotVM.PUT("/target-region", spotVMHandler.SetTargetRegion)
				spotVM.POST("/trigger-termination", spotVMHandler.TriggerTermination)

				// 自动管理相关路由
				spotVM.GET("/auto-manager/status", spotVMHandler.GetAutoManagerStatus)
				spotVM.POST("/auto-manager/start", spotVMHandler.StartAutoManager)
				spotVM.POST("/auto-manager/stop", spotVMHandler.StopAutoManager)
				spotVM.POST("/auto-manager/simulate-termination", spotVMHandler.SimulateTermination)
			}

			// 示例：获取实例信息（保留原有功能）
			protected.GET("/instances", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Instance information",
					"data":    []string{"instance1", "instance2"},
				})
			})

			// 示例：创建实例（保留原有功能）
			protected.POST("/instances", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message":     "Instance created successfully",
					"instance_id": "ins-123456",
				})
			})
		}
	}

	return router
}
