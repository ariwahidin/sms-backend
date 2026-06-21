package handler

import (
	"strconv"

	"sms-backend/service"

	"github.com/gofiber/fiber/v2"
)

type LocationHandler struct {
	locationSvc *service.LocationService
}

func NewLocationHandler(locationSvc *service.LocationService) *LocationHandler {
	return &LocationHandler{locationSvc: locationSvc}
}

func (h *LocationHandler) GetAll(c *fiber.Ctx) error {
	locations, err := h.locationSvc.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(locations)
}

func (h *LocationHandler) GetByQRToken(c *fiber.Ctx) error {
	token := c.Params("qr_token")
	loc, err := h.locationSvc.GetByQRToken(token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}
	return c.JSON(loc)
}

func (h *LocationHandler) Create(c *fiber.Ctx) error {
	var input service.CreateLocationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	loc, err := h.locationSvc.Create(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(loc)
}

func (h *LocationHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var input service.UpdateLocationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	loc, err := h.locationSvc.Update(uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(loc)
}
