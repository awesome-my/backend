package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

func (c *Client) Events(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)
	page, limit, offset := awesomemy.PageLimitOffsetFromRequest(r)
	orderBy := "desc"
	if r.URL.Query().Get("orderBy") == "asc" {
		orderBy = "asc"
	}

	var events []database.Event
	var err error
	switch orderBy {
	case "asc":
		events, err = c.queries.UserEventsByAscOffsetLimit(r.Context(), c.database, database.UserEventsByAscOffsetLimitParams{
			Offset: int32(offset),
			Limit:  int32(limit),
			UserID: authUser.UserID,
		})
	case "desc":
		events, err = c.queries.UserEventsByDescOffsetLimit(r.Context(), c.database, database.UserEventsByDescOffsetLimitParams{
			Offset: int32(offset),
			Limit:  int32(limit),
			UserID: authUser.UserID,
		})
	}
	if err != nil {
		c.logger.Error("could not fetch user events by limit offset", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch user events.",
		})
		return
	}

	total, err := c.queries.CountUserEvents(r.Context(), c.database, authUser.UserID)
	if err != nil {
		c.logger.Error("could not fetch user events count", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch user events count.",
		})
		return
	}

	apiEvents := make([]Event, len(events))
	for i, e := range events {
		apiEvents[i] = EventFromDatabase(e)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"items":      apiEvents,
		"pagination": awesomemy.NewPaginationMeta(page, len(events), int(total)),
	})
}

func (c *Client) Event(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	eventUuid, err := uuid.FromString(chi.URLParam(r, "event"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	event, err := c.queries.EventByUUID(r.Context(), c.database, eventUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		c.logger.Error("could not fetch event by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch event.",
		})
		return
	}

	if event.UserID != authUser.UserID {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": EventFromDatabase(event),
	})
}

func (c *Client) StoreEvent(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	var data struct {
		Name        string    `json:"name" validate:"required,min=8,max=191"`
		Description string    `json:"description" validate:"required,min=8,max=512"`
		Tags        []string  `json:"tags" validate:"min=0,max=6,dive,min=4,max=12"`
		Website     string    `json:"website" validate:"omitempty,url,max=191"`
		StartsAt    time.Time `json:"starts_at" validate:"required"`
		EndsAt      time.Time `json:"ends_at" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	if err := c.validator.StructCtx(r.Context(), data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	count, err := c.queries.CountUserEvents(r.Context(), c.database, authUser.UserID)
	if err != nil {
		c.logger.Error("could not fetch user events count", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch user events count.",
		})
		return
	}

	if count >= 20 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "You have hit the event limit, try deleting some unused events.",
		})
		return
	}

	var website nulls.String
	if data.Website != "" {
		website = nulls.String{
			String: data.Website,
			Valid:  true,
		}
	}

	event, err := c.queries.InsertEvent(r.Context(), c.database, database.InsertEventParams{
		Name:        data.Name,
		Description: data.Description,
		Tags:        data.Tags,
		Website:     website,
		StartsAt:    data.StartsAt,
		EndsAt:      data.EndsAt,
		UserID:      authUser.UserID,
	})
	if err != nil {
		c.logger.Error("could not insert event", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not insert event into database.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": EventFromDatabase(event),
	})
}

func (c *Client) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	eventUuid, err := uuid.FromString(chi.URLParam(r, "event"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	event, err := c.queries.EventByUUID(r.Context(), c.database, eventUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		c.logger.Error("could not fetch event by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch event.",
		})
		return
	}

	if event.UserID != authUser.UserID {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	var data struct {
		Name        string    `json:"name" validate:"required,min=8,max=191"`
		Description string    `json:"description" validate:"required,min=8,max=512"`
		Tags        []string  `json:"tags" validate:"min=0,max=6,dive,min=4,max=12"`
		Website     string    `json:"website" validate:"omitempty,url,max=191"`
		StartsAt    time.Time `json:"starts_at" validate:"required"`
		EndsAt      time.Time `json:"ends_at" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	if err := c.validator.StructCtx(r.Context(), data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	var website nulls.String
	if data.Website != "" {
		website = nulls.String{
			String: data.Website,
			Valid:  true,
		}
	}

	event, err = c.queries.UpdateEvent(r.Context(), c.database, database.UpdateEventParams{
		Name:        data.Name,
		Description: data.Description,
		Tags:        data.Tags,
		Website:     website,
		StartsAt:    data.StartsAt,
		EndsAt:      data.EndsAt,
		EventID:     event.EventID,
	})
	if err != nil {
		c.logger.Error("could not update event", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not update event.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": EventFromDatabase(event),
	})
}

func (c *Client) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	eventUuid, err := uuid.FromString(chi.URLParam(r, "event"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	event, err := c.queries.EventByUUID(r.Context(), c.database, eventUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		c.logger.Error("could not fetch event by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch event.",
		})
		return
	}

	if event.UserID != authUser.UserID {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	if err := c.queries.DeleteEvent(r.Context(), c.database, event.EventID); err != nil {
		c.logger.Error("could not delete event", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not delete event.",
		})
		return
	}
}
