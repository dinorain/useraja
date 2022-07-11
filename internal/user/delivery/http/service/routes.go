package http

func (h *userHandlersHTTP) AdminMapRoutes() {
	h.group.POST("/refresh", h.RefreshToken())
	h.group.POST("/login", h.Login())
	h.group.POST("/logout", h.Logout(), h.mw.IsLoggedIn())

	h.group.Use(h.mw.IsLoggedIn(), h.mw.IsAdmin)
	h.group.POST("", h.Register())

	h.group.PUT("/me", h.UpdateByID(), h.mw.IsLoggedIn())
	h.group.GET("/me", h.GetMe(), h.mw.IsLoggedIn())

	h.group.GET("/:id", h.FindByID())
	h.group.PUT("/:id", h.UpdateByID())
	h.group.DELETE("/:id", h.DeleteByID())
}

func (h *userHandlersHTTP) UserMapRoutes() {
	h.group.POST("/refresh", h.RefreshToken())
	h.group.POST("/login", h.Login())
	h.group.POST("/logout", h.Logout(), h.mw.IsLoggedIn())

	h.group.PUT("/me", h.UpdateByID(), h.mw.IsLoggedIn())
	h.group.GET("/me", h.GetMe(), h.mw.IsLoggedIn())
}
