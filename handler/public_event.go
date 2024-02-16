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

type Event struct {
	Uuid        uuid.UUID    `json:"uuid"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	Website     nulls.String `json:"website"`
	StartsAt    time.Time    `json:"starts_at"`
	EndsAt      time.Time    `json:"ends_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

func EventFromDatabase(e database.Event) Event {
	return Event{
		Uuid:        e.Uuid,
		Name:        e.Name,
		Description: e.Description,
		Tags:        e.Tags,
		Website:     e.Website,
		StartsAt:    e.StartsAt,
		EndsAt:      e.EndsAt,
		CreatedAt:   e.CreatedAt,
	}
}

func (p *Public) Events(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := awesomemy.PageLimitOffsetFromRequest(r)

	events, err := p.queries.EventsByOffsetLimit(r.Context(), p.database, database.EventsByOffsetLimitParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		p.logger.Error("could not fetch events by limit offset", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch events.",
		})
		return
	}

	total, err := p.queries.CountEvents(r.Context(), p.database)
	if err != nil {
		p.logger.Error("could not fetch events count", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch events count.",
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

func (p *Public) Event(w http.ResponseWriter, r *http.Request) {
	eventUuid, err := uuid.FromString(chi.URLParam(r, "event"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	event, err := p.queries.EventByUUID(r.Context(), p.database, eventUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		p.logger.Error("could not fetch event by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch event.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": EventFromDatabase(event),
	})
}
