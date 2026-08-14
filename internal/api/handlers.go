package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jeeftor/bookSync/internal/matcher"
	"github.com/jeeftor/bookSync/internal/store"
)

// Health reports server status and build info, used by the frontend (to show
// the running version) and for monitoring/liveness checks.
func (h *Handlers) Health(build BuildInfo) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"status": "ok", "version": build.Version, "commit": build.Commit, "date": build.Date})
	}
}

func idParam(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func fail(c echo.Context, status int, err error) error {
	return c.JSON(status, map[string]string{"error": err.Error()})
}

// --- Kindle accounts ---------------------------------------------------

func (h *Handlers) ListKindleAccounts(c echo.Context) error {
	accs, err := h.svc.ListKindleAccounts(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, accs)
}

func (h *Handlers) CreateKindleAccount(c echo.Context) error {
	var in store.KindleAccount
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateKindleAccount(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handlers) UpdateKindleAccount(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	var in store.KindleAccount
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	in.ID = id
	out, err := h.svc.UpdateKindleAccount(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handlers) DeleteKindleAccount(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.DeleteKindleAccount(c.Request().Context(), id); err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) TestKindleAccount(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	count, err := h.svc.TestKindleAccount(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusBadGateway, err)
	}
	return c.JSON(http.StatusOK, map[string]int{"bookCount": count})
}

// --- Audiobookshelf users -----------------------------------------------

func (h *Handlers) ListABSUsers(c echo.Context) error {
	users, err := h.svc.ListABSUsers(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, users)
}

func (h *Handlers) CreateABSUser(c echo.Context) error {
	var in store.ABSUser
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateABSUser(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handlers) UpdateABSUser(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	var in store.ABSUser
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	in.ID = id
	out, err := h.svc.UpdateABSUser(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handlers) DeleteABSUser(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.DeleteABSUser(c.Request().Context(), id); err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) TestABSUser(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	libs, err := h.svc.TestABSUser(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusBadGateway, err)
	}
	return c.JSON(http.StatusOK, libs)
}

// --- Profiles -------------------------------------------------------------

func (h *Handlers) ListProfiles(c echo.Context) error {
	profiles, err := h.svc.ListProfiles(c.Request().Context())
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, profiles)
}

func (h *Handlers) CreateProfile(c echo.Context) error {
	var in store.Profile
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	out, err := h.svc.CreateProfile(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handlers) UpdateProfile(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	var in store.Profile
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	in.ID = id
	out, err := h.svc.UpdateProfile(c.Request().Context(), in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handlers) DeleteProfile(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.DeleteProfile(c.Request().Context(), id); err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) Suggestions(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	candidates, err := h.svc.Suggestions(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusBadGateway, err)
	}
	return c.JSON(http.StatusOK, candidates)
}

func (h *Handlers) ListMappings(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	mappings, err := h.svc.ListMappings(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, mappings)
}

func (h *Handlers) ConfirmMatch(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	var in matcher.Candidate
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	m, err := h.svc.ConfirmMatch(c.Request().Context(), id, in)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handlers) RejectMatch(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	var in matcher.Candidate
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.RejectMatch(c.Request().Context(), id, in); err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) SyncProfile(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.SyncProfile(c.Request().Context(), id); err != nil {
		return fail(c, http.StatusBadGateway, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Mappings ---------------------------------------------------------

func (h *Handlers) DeleteMapping(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	if err := h.svc.DeleteMapping(c.Request().Context(), id); err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) SyncMapping(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	event, err := h.svc.SyncMapping(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusBadGateway, err)
	}
	return c.JSON(http.StatusOK, event)
}

func (h *Handlers) MappingHistory(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, err)
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	events, err := h.svc.SyncHistory(c.Request().Context(), id, limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, events)
}

func (h *Handlers) RecentActivity(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	events, err := h.svc.RecentActivity(c.Request().Context(), limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, events)
}
