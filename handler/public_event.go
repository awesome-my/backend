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
		Website:     e.Website,
		StartsAt:    e.StartsAt,
		EndsAt:      e.EndsAt,
		CreatedAt:   e.CreatedAt,
	}
}

func (p *Public) Events(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := awesomemy.PageLimitOffsetFromRequest(r)

	//var tags []string
	//if r.URL.Query().Get("tags") != "" {
	//	tags = strings.Split(r.URL.Query().Get("tags"), ",")
	//}

	keyword := r.URL.Query().Get("keyword")
	orderBy := "desc"
	if r.URL.Query().Get("orderBy") == "asc" {
		orderBy = "asc"
	}

	var err error
	var events []database.Event
	switch orderBy {
	case "asc":
		if len(tags) > 0 {
			events, err = p.queries.EventsByTagsAscOffsetLimit(r.Context(), p.database, database.EventsByTagsAscOffsetLimitParams{
				Tags:   tags,
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		} else {
			events, err = p.queries.EventsByAscOffsetLimit(r.Context(), p.database, database.EventsByAscOffsetLimitParams{
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		}
	case "desc":
		if len(tags) > 0 {
			events, err = p.queries.EventsByTagsDescOffsetLimit(r.Context(), p.database, database.EventsByTagsDescOffsetLimitParams{
				Tags:   tags,
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		} else {
			events, err = p.queries.EventsByDescOffsetLimit(r.Context(), p.database, database.EventsByDescOffsetLimitParams{
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		}
	}
	if err != nil {
		p.logger.Error("could not fetch events by limit offset", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch events.",
		})
		return
	}

	var total int64
	if len(tags) > 0 {
		total, err = p.queries.CountEventsByTags(r.Context(), p.database, tags)
	} else {
		total, err = p.queries.CountEvents(r.Context(), p.database)
	}
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
