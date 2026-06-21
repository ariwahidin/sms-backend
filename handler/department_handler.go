package handler

import (
	"strconv"

	"sms-backend/service"

	"github.com/gofiber/fiber/v2"
)

type DepartmentHandler struct {
	deptSvc *service.DepartmentService
}

func NewDepartmentHandler(deptSvc *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptSvc: deptSvc}
}

func (h *DepartmentHandler) GetAll(c *fiber.Ctx) error {
	depts, err := h.deptSvc.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(depts)
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	var input service.CreateDepartmentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dept, err := h.deptSvc.Create(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(dept)
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var input service.UpdateDepartmentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dept, err := h.deptSvc.Update(uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dept)
}
