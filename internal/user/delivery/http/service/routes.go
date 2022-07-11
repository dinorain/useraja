package http

func (h *userHandlersHTTP) AdminMapRoutes() {
	h.group.Group("", h.mw.IsLoggedIn(), h.mw.IsAdmin)
	h.group.POST("", h.Register())
	h.group.PUT("/:id", h.UpdateByID())
}

func (h *userHandlersHTTP) UserMapRoutes() {
	h.group.PUT("/:id", h.UpdateByID(), h.mw.IsLoggedIn())

	h.group.GET("/me", h.GetMe(), h.mw.IsLoggedIn())
	h.group.GET("/:id", h.FindByID())

	h.group.POST("/login", h.Login())
	h.group.POST("/logout", h.Logout(), h.mw.IsLoggedIn())
}
