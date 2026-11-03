package handler

import (
	"errors"
	"strconv"

	"sms-backend/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SafetyCategoryHandler struct {
	service *service.SafetyCategoryService
}

func NewSafetyCategoryHandler(
	svc *service.SafetyCategoryService,
) *SafetyCategoryHandler {
	return &SafetyCategoryHandler{
		service: svc,
	}
}

func (h *SafetyCategoryHandler) GetAll(c *fiber.Ctx) error {
	categories, err := h.service.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve safety categories",
		})
	}

	return c.JSON(fiber.Map{
		"data":  categories,
		"total": len(categories),
	})
}

func (h *SafetyCategoryHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid safety category ID",
		})
	}

	category, err := h.service.GetByID(uint(id))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(category)
}

func (h *SafetyCategoryHandler) Create(c *fiber.Ctx) error {
	var input service.CreateSafetyCategoryInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category, err := h.service.Create(input)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}

func (h *SafetyCategoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid safety category ID",
		})
	}

	var input service.UpdateSafetyCategoryInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category, err := h.service.Update(uint(id), input)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(category)
}

func (h *SafetyCategoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid safety category ID",
		})
	}

	if err := h.service.Delete(uint(id)); err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "Safety category deactivated successfully",
	})
}

func (h *SafetyCategoryHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrSafetyCategoryNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})

	case errors.Is(err, service.ErrSafetyCategoryCodeUsed):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})

	case errors.Is(err, service.ErrSafetyCategoryCodeRequired),
		errors.Is(err, service.ErrSafetyCategoryNameRequired),
		errors.Is(err, service.ErrSafetyCategoryInactive):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
}

func (h *SafetyCategoryHandler) GetPublic(c *fiber.Ctx) error {
	categories, err := h.service.GetActive()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve safety categories",
		})
	}

	return c.JSON(fiber.Map{
		"data":  categories,
		"total": len(categories),
	})
}
