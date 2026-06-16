package handler

import (
	"github.com/gofiber/fiber/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

// portalJamaah pulls the scoped jamaah id from the token. RequireJamaahScope
// guarantees it is non-nil.
func portalJamaah(c *fiber.Ctx) (claims *sharedAuth.Claims, ok bool) {
	claims = c.Locals("claims").(*sharedAuth.Claims)
	if claims.JamaahID == nil {
		return nil, false
	}
	return claims, true
}

func (h *JamaahHandler) PortalProfile(c *fiber.Ctx) error {
	claims, ok := portalJamaah(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	p, err := h.svc.PortalProfile(c.Context(), claims.OrgID, *claims.JamaahID)
	if err != nil {
		return response.NotFound(c, "profil tidak ditemukan")
	}
	return response.OK(c, p)
}

func (h *JamaahHandler) PortalRegistrations(c *fiber.Ctx) error {
	claims, ok := portalJamaah(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	regs, err := h.svc.PortalRegistrations(c.Context(), claims.OrgID, *claims.JamaahID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"registrations": regs})
}

func (h *JamaahHandler) PortalDocuments(c *fiber.Ctx) error {
	claims, ok := portalJamaah(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	docs, err := h.svc.PortalDocuments(c.Context(), claims.OrgID, *claims.JamaahID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"documents": docs})
}

func (h *JamaahHandler) PortalVisa(c *fiber.Ctx) error {
	claims, ok := portalJamaah(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	v, err := h.svc.PortalVisa(c.Context(), claims.OrgID, *claims.JamaahID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, v)
}

func (h *JamaahHandler) PortalPayments(c *fiber.Ctx) error {
	claims, ok := portalJamaah(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	bal := h.svc.PortalPayments(c.Context(), c.Get("Authorization"), *claims.JamaahID)
	return response.OK(c, bal)
}
