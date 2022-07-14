package handlers

func (h *userHandlersHTTP) UserMapRoutes() {
	h.group.POST("/refresh", h.RefreshToken())
	h.group.POST("/login", h.Login())

	h.group.Use(h.mw.IsLoggedIn())
	h.group.POST("/logout", h.Logout())
	h.group.GET("/:id", h.FindById())
	h.group.PUT("/:id", h.UpdateById())
	h.group.GET("/me", h.GetMe())

	h.group.GET("", h.FindAll())
	h.group.POST("", h.Register(), h.mw.IsAdmin)
	h.group.DELETE("/:id", h.DeleteById(), h.mw.IsAdmin)
}
